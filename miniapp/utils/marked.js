/**
 * 轻量级 Markdown → HTML 渲染器
 * 替代 marked.js，减小包体积，适配小程序环境
 */

function escapeHtml(text) {
  return text
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
}

function parseInline(text) {
  return text
    .replace(/\*\*(.+?)\*\*/g, '<strong>$1</strong>')
    .replace(/\*(.+?)\*/g, '<em>$1</em>')
    .replace(/`(.+?)`/g, '<code style="background:#f0ebe0;padding:2rpx 8rpx;border-radius:4rpx;font-size:0.9em;">$1</code>')
    .replace(/~~(.+?)~~/g, '<del>$1</del>')
}

export default {
  parse(text) {
    if (!text) return ''

    const lines = text.split('\n')
    const html = []
    let inCodeBlock = false
    let codeBuffer = []
    let inList = false

    for (let i = 0; i < lines.length; i++) {
      const line = lines[i]

      // 代码块 ``` ... ```
      if (line.trimStart().startsWith('```')) {
        if (inCodeBlock) {
          html.push('<pre style="background:#f0ebe0;padding:16rpx;border-radius:8rpx;overflow-x:auto;font-size:24rpx;"><code>' + escapeHtml(codeBuffer.join('\n')) + '</code></pre>')
          codeBuffer = []
          inCodeBlock = false
        } else {
          inCodeBlock = true
          codeBuffer = []
        }
        continue
      }
      if (inCodeBlock) {
        codeBuffer.push(line)
        continue
      }

      // 空行
      if (line.trim() === '') {
        if (inList) {
          html.push('</ul>')
          inList = false
        }
        continue
      }

      // 水平线
      if (/^-{3,}$/.test(line.trim()) || /^\*{3,}$/.test(line.trim())) {
        html.push('<hr style="border:none;border-top:1rpx solid #E0D6C8;margin:24rpx 0;" />')
        continue
      }

      // 标题
      const headingMatch = line.match(/^(#{1,6})\s+(.+)$/)
      if (headingMatch) {
        const level = headingMatch[1].length
        const content = parseInline(escapeHtml(headingMatch[2]))
        const sizes = { 1: '36rpx', 2: '32rpx', 3: '30rpx', 4: '28rpx', 5: '26rpx', 6: '24rpx' }
        html.push(`<h${level} style="font-size:${sizes[level]||'28rpx'};font-weight:700;margin:20rpx 0 12rpx;color:#3D3226;">${content}</h${level}>`)
        if (inList) { html.push('</ul>'); inList = false }
        continue
      }

      // 有序列表
      const olMatch = line.match(/^\d+[.)]\s+(.+)$/)
      if (olMatch) {
        if (!inList) { html.push('<ol style="padding-left:40rpx;margin:8rpx 0;">'); inList = true }
        html.push(`<li style="margin:4rpx 0;">${parseInline(escapeHtml(olMatch[1]))}</li>`)
        continue
      }

      // 无序列表
      const ulMatch = line.match(/^[-*+]\s+(.+)$/)
      if (ulMatch) {
        if (!inList) { html.push('<ul style="padding-left:40rpx;margin:8rpx 0;">'); inList = true }
        html.push(`<li style="margin:4rpx 0;">${parseInline(escapeHtml(ulMatch[1]))}</li>`)
        continue
      }

      // 闭合列表
      if (inList) {
        // 非列表行 — 视为续行
        html[html.length - 1] = html[html.length - 1].replace('</li>', '') + '<br/>' + parseInline(escapeHtml(line)) + '</li>'
        continue
      }

      // 引用块
      const bqMatch = line.match(/^>\s*(.*)$/)
      if (bqMatch) {
        html.push(`<blockquote style="border-left:4rpx solid #A0522D;padding:12rpx 16rpx;margin:12rpx 0;background:rgba(160,82,45,0.05);border-radius:0 8rpx 8rpx 0;color:#5D4037;">${parseInline(escapeHtml(bqMatch[1]))}</blockquote>`)
        continue
      }

      // 普通段落
      html.push(`<p style="margin:8rpx 0;line-height:1.8;">${parseInline(escapeHtml(line))}</p>`)
    }

    if (inCodeBlock) {
      html.push('<pre style="background:#f0ebe0;padding:16rpx;border-radius:8rpx;overflow-x:auto;font-size:24rpx;"><code>' + escapeHtml(codeBuffer.join('\n')) + '</code></pre>')
    }
    if (inList) {
      html.push('</ul>')
    }

    return html.join('\n')
  }
}
