<template>
  <div class="markdown-body" v-html="rendered"></div>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({ content: { type: String, default: '' } })

// 简易 Markdown 渲染器（处理解卦常用格式）
const rendered = computed(() => {
  let html = props.content
    // 代码块
    .replace(/```(\w*)\n([\s\S]*?)```/g, '<pre class="code-block"><code>$2</code></pre>')
    // 行内代码
    .replace(/`([^`]+)`/g, '<code class="inline-code">$1</code>')
    // 标题
    .replace(/^#### (.+)$/gm, '<h4>$1</h4>')
    .replace(/^### (.+)$/gm, '<h3>$1</h3>')
    .replace(/^## (.+)$/gm, '<h2>$1</h2>')
    .replace(/^# (.+)$/gm, '<h1>$1</h1>')
    // 引用
    .replace(/^> (.+)$/gm, '<blockquote>$1</blockquote>')
    // 粗体
    .replace(/\*\*(.+?)\*\*/g, '<strong>$1</strong>')
    // 斜体
    .replace(/\*(.+?)\*/g, '<em>$1</em>')
    // 无序列表
    .replace(/^- (.+)$/gm, '<li>$1</li>')
    .replace(/(<li>.*<\/li>\n?)+/g, '<ul>$&</ul>')
    // 有序列表
    .replace(/^\d+\. (.+)$/gm, '<li>$1</li>')
    // 换行转段落
    .replace(/\n\n/g, '</p><p>')
    .replace(/^(.+)$/gm, (m) => {
      if (m.startsWith('<')) return m
      return m
    })

  // 包裹段落
  if (!html.startsWith('<')) {
    html = '<p>' + html + '</p>'
  }
  // 修复嵌套 p
  html = html.replace('</p><p><p>', '</p><p>').replace('</p></p>', '</p>')

  return html
})
</script>

<style scoped>
.markdown-body {
  font-size: 14px;
  line-height: 1.8;
  color: #1c1917;
}

.markdown-body :deep(h2) {
  font-size: 18px;
  font-weight: 700;
  margin: 20px 0 8px;
  color: #1c1917;
}

.markdown-body :deep(h3) {
  font-size: 16px;
  font-weight: 600;
  margin: 16px 0 6px;
}

.markdown-body :deep(h4) {
  font-size: 14px;
  font-weight: 600;
  margin: 12px 0 4px;
}

.markdown-body :deep(p) {
  margin: 8px 0;
}

.markdown-body :deep(strong) {
  font-weight: 700;
  color: #d4a853;
}

.markdown-body :deep(blockquote) {
  border-left: 3px solid #d4a853;
  padding: 8px 16px;
  margin: 12px 0;
  background: #fefaf0;
  border-radius: 0 6px 6px 0;
  color: #78716c;
  font-style: italic;
}

.markdown-body :deep(ul), .markdown-body :deep(ol) {
  padding-left: 20px;
  margin: 8px 0;
}

.markdown-body :deep(li) {
  margin: 4px 0;
}

.markdown-body :deep(.code-block) {
  background: #1a1a2e;
  color: #e8e4d8;
  padding: 12px 16px;
  border-radius: 8px;
  overflow-x: auto;
  font-size: 13px;
  line-height: 1.5;
}

.markdown-body :deep(.inline-code) {
  background: #fefaf0;
  color: #b8860b;
  padding: 1px 6px;
  border-radius: 4px;
  font-size: 13px;
}

.markdown-body :deep(em) {
  font-style: italic;
  color: #78716c;
}
</style>
