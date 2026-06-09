from fastapi import APIRouter, Query
from app.domain.seeds.market_seed_loader import MarketSeedLoader
from app.domain.predict.predictor import StockLGBMPredictor
from app.settings import get_settings

router = APIRouter(prefix="/internal/v1", tags=["predict"])


@router.get("/predict/stock-7d")
def predict_stock_7d(
    symbol: str = Query(..., description="Stock symbol"),
    trade_date: str = Query(..., description="Trade date YYYY-MM-DD"),
):
    settings = get_settings()
    loader = MarketSeedLoader(settings=settings)
    
    # Load 120 days of feature history
    seeds = loader.load_history(symbol=symbol, trade_date=trade_date, limit=120)
    
    # Train model and predict
    predictor = StockLGBMPredictor()
    prediction = predictor.predict_next_7d(seeds)
    
    return prediction
