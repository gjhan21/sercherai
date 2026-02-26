<script setup>
import { nextTick, onMounted, ref, watch } from "vue";

const props = defineProps({
  modelValue: {
    type: String,
    default: ""
  },
  placeholder: {
    type: String,
    default: "请输入内容"
  },
  minHeight: {
    type: Number,
    default: 260
  }
});

const emit = defineEmits(["update:modelValue"]);

const editorRef = ref(null);

function getHTML() {
  return editorRef.value?.innerHTML || "";
}

function setHTML(value) {
  if (!editorRef.value) {
    return;
  }
  const normalized = value || "";
  if (editorRef.value.innerHTML !== normalized) {
    editorRef.value.innerHTML = normalized;
  }
  updateEmptyState();
}

function updateEmptyState() {
  if (!editorRef.value) {
    return;
  }
  const text = (editorRef.value.textContent || "").replace(/\u200B/g, "").trim();
  editorRef.value.dataset.empty = text ? "false" : "true";
}

function emitValue() {
  emit("update:modelValue", getHTML());
  updateEmptyState();
}

function focusEditor() {
  editorRef.value?.focus();
}

function exec(command, value = null) {
  focusEditor();
  document.execCommand(command, false, value);
  emitValue();
}

function insertLink() {
  const url = window.prompt("请输入链接地址", "https://");
  if (!url) {
    return;
  }
  exec("createLink", url.trim());
}

function clearFormat() {
  exec("removeFormat");
  exec("unlink");
}

function handlePaste(event) {
  if (!event.clipboardData) {
    return;
  }
  const text = event.clipboardData.getData("text/plain");
  if (!text) {
    return;
  }
  event.preventDefault();
  exec("insertText", text);
}

watch(
  () => props.modelValue,
  (value) => {
    if (!editorRef.value) {
      return;
    }
    const isFocused = document.activeElement === editorRef.value;
    if (isFocused) {
      return;
    }
    setHTML(value);
  }
);

onMounted(async () => {
  await nextTick();
  setHTML(props.modelValue);
});
</script>

<template>
  <div class="rte-shell">
    <div class="rte-toolbar">
      <el-button-group>
        <el-tooltip content="加粗">
          <el-button text size="small" @click="exec('bold')"><b>B</b></el-button>
        </el-tooltip>
        <el-tooltip content="斜体">
          <el-button text size="small" @click="exec('italic')"><i>I</i></el-button>
        </el-tooltip>
        <el-tooltip content="下划线">
          <el-button text size="small" @click="exec('underline')"><u>U</u></el-button>
        </el-tooltip>
        <el-tooltip content="标题">
          <el-button text size="small" @click="exec('formatBlock', 'h3')">H3</el-button>
        </el-tooltip>
        <el-tooltip content="无序列表">
          <el-button text size="small" @click="exec('insertUnorderedList')">列表</el-button>
        </el-tooltip>
      </el-button-group>
      <div class="toolbar-spacer" />
      <el-button text size="small" @click="insertLink">插入链接</el-button>
      <el-button text size="small" @click="clearFormat">清除格式</el-button>
    </div>

    <div
      ref="editorRef"
      class="rte-editor"
      contenteditable="true"
      :data-placeholder="placeholder"
      :style="{ minHeight: `${minHeight}px` }"
      @input="emitValue"
      @blur="emitValue"
      @paste="handlePaste"
    />
  </div>
</template>

<style scoped>
.rte-shell {
  border: 1px solid #d1d5db;
  border-radius: 10px;
  overflow: hidden;
  background: #fff;
}

.rte-toolbar {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px;
  border-bottom: 1px solid #e5e7eb;
  background: #f8fafc;
}

.toolbar-spacer {
  flex: 1;
}

.rte-editor {
  padding: 12px;
  outline: none;
  line-height: 1.6;
  overflow-y: auto;
}

.rte-editor[data-empty="true"]::before {
  content: attr(data-placeholder);
  color: #9ca3af;
}

.rte-editor h3 {
  margin: 8px 0;
}

.rte-editor :deep(ul) {
  margin: 8px 0;
  padding-left: 18px;
}
</style>
