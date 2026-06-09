import numpy as np
import pandas as pd
import lightgbm as lgb
from typing import Any, Dict, List
from app.domain.models import MarketSeed

class StockLGBMPredictor:
    FEATURES = [
        "momentum1", "momentum2", "momentum3", "momentum5", "momentum20",
        "volatility20", "volume_ratio", "drawdown20", "trend_strength",
        "net_mf_amount", "pe_ttm", "pb", "turnover_rate", "news_heat", "positive_news_rate",
        "deviation_ma5", "deviation_ma10", "deviation_ma20", "deviation_ma60",
        "candle_body_pct", "lower_shadow_pct", "upper_shadow_pct"
    ]

    def predict_next_7d(self, seeds: List[MarketSeed]) -> Dict[str, Any]:
        """
        在个股的历史特征序列上动态训练 LightGBM 模型，并预测未来 7 天的走势区间。
        """
        if not seeds or len(seeds) < 20:
            return self._build_empty_prediction()

        data_dict = []
        for s in seeds:
            row = {
                "date": s.trade_date,
                "close": s.close_price,
            }
            for feat in self.FEATURES:
                row[feat] = getattr(s, feat, 0.0) or 0.0
            data_dict.append(row)

        df = pd.DataFrame(data_dict)
        df = df.sort_values("date").reset_index(drop=True)
        n_samples = len(df)

        if n_samples < 20:
            return self._build_empty_prediction()

        # 1. 准备特征集
        X = df[self.FEATURES].fillna(0.0)
        X_latest = X.iloc[[-1]]

        # 计算历史日均波动率并约束在 [0.5, 6.0] 之间
        closes = df["close"].values
        daily_returns = np.diff(closes) / closes[:-1] * 100.0 if len(closes) > 1 else np.array([])
        volatility = float(np.std(daily_returns)) if len(daily_returns) > 1 else 1.5
        if np.isnan(volatility) or volatility <= 0:
            volatility = 1.5
        volatility = max(0.5, min(6.0, volatility))

        medians = []
        upper_75 = []
        lower_25 = []

        # 2. 为未来 1..7 天分别进行分位数回归训练
        for d in range(1, 8):
            y_d = []
            valid_idx = []
            for i in range(n_samples):
                if i + d < n_samples:
                    ret = (df.loc[i + d, "close"] / df.loc[i, "close"] - 1.0) * 100.0
                    y_d.append(ret)
                    valid_idx.append(i)

            if len(y_d) < 10:
                # 样本数过少时回退
                medians.append(0.0)
                upper_75.append(0.0)
                lower_25.append(0.0)
                continue

            X_train = X.iloc[valid_idx]
            y_train = np.array(y_d)

            # 训练中位数模型 (50分位数)
            model_med = lgb.LGBMRegressor(
                objective="quantile",
                alpha=0.50,
                n_estimators=35,
                max_depth=3,
                num_leaves=7,
                learning_rate=0.1,
                min_child_samples=5,
                random_state=42,
                verbose=-1
            )
            model_med.fit(X_train, y_train)
            pred_med = float(model_med.predict(X_latest)[0])

            # 训练上限模型 (75分位数)
            model_up = lgb.LGBMRegressor(
                objective="quantile",
                alpha=0.75,
                n_estimators=35,
                max_depth=3,
                num_leaves=7,
                learning_rate=0.1,
                min_child_samples=5,
                random_state=42,
                verbose=-1
            )
            model_up.fit(X_train, y_train)
            pred_up = float(model_up.predict(X_latest)[0])

            # 训练下限模型 (25分位数)
            model_low = lgb.LGBMRegressor(
                objective="quantile",
                alpha=0.25,
                n_estimators=35,
                max_depth=3,
                num_leaves=7,
                learning_rate=0.1,
                min_child_samples=5,
                random_state=42,
                verbose=-1
            )
            model_low.fit(X_train, y_train)
            pred_low = float(model_low.predict(X_latest)[0])

            # 防止分位数穿越 (Quantile Crossing)
            pred_low_final = min(pred_low, pred_med, pred_up)
            pred_up_final = max(pred_low, pred_med, pred_up)
            pred_med_final = pred_med
            if pred_med_final < pred_low_final:
                pred_med_final = pred_low_final
            elif pred_med_final > pred_up_final:
                pred_med_final = pred_up_final

            medians.append(round(pred_med_final, 4))
            upper_75.append(round(pred_up_final, 4))
            lower_25.append(round(pred_low_final, 4))

        # 3. 注入布朗桥随机噪声，增加 K 线真实波动质感
        # 根据股票 symbol 和当前交易日期生成唯一的随机种子，确保不同股票拥有个性化的波动路径
        symbol_str = seeds[0].symbol if seeds else "default"
        date_str = seeds[-1].trade_date if seeds else "today"
        import hashlib
        seed_hash = hashlib.md5(f"{symbol_str}_{date_str}".encode("utf-8")).hexdigest()
        seed_int = int(seed_hash[:8], 16)
        np.random.seed(seed_int)
        raw_noise = np.random.normal(0, 0.6 * volatility, 7)
        mean_noise = np.mean(raw_noise)
        centered_noise = raw_noise - mean_noise
        cum_noise = np.cumsum(centered_noise)

        noisy_medians = []
        noisy_upper_75 = []
        noisy_lower_25 = []

        for d in range(7):
            lgb_med = medians[d]
            noise = cum_noise[d]
            new_med = lgb_med + noise
            
            # 保持原置信通道宽度不变并叠加
            half_width_up = max(0.0, upper_75[d] - lgb_med)
            half_width_low = max(0.0, lgb_med - lower_25[d])
            
            new_up = new_med + half_width_up
            new_low = new_med - half_width_low
            
            # 防止分位数穿越
            new_low_final = min(new_low, new_med, new_up)
            new_up_final = max(new_low, new_med, new_up)
            new_med_final = new_med
            if new_med_final < new_low_final:
                new_med_final = new_low_final
            elif new_med_final > new_up_final:
                new_med_final = new_up_final

            noisy_medians.append(round(new_med_final, 4))
            noisy_upper_75.append(round(new_up_final, 4))
            noisy_lower_25.append(round(new_low_final, 4))

        # 4. 计算统计摘要 (使用叠加噪声后的预测中位数)
        avg_return = round(float(np.mean(noisy_medians)), 2)

        # 依据历史回测数据计算胜率 (7日涨幅大于0的样本比例)
        y_7d_train = []
        for i in range(n_samples):
            if i + 7 < n_samples:
                y_7d_train.append((df.loc[i + 7, "close"] / df.loc[i, "close"] - 1.0) * 100.0)

        if y_7d_train:
            win_rate = round(float(np.sum(np.array(y_7d_train) > 0) / len(y_7d_train) * 100.0), 2)
        else:
            win_rate = 50.0

        return {
            "median": noisy_medians,
            "upper_75": noisy_upper_75,
            "lower_25": noisy_lower_25,
            "avg_return": avg_return,
            "win_rate": win_rate
        }

    def _build_empty_prediction(self) -> Dict[str, Any]:
        return {
            "median": [0.0] * 7,
            "upper_75": [0.0] * 7,
            "lower_25": [0.0] * 7,
            "avg_return": 0.0,
            "win_rate": 50.0
        }
