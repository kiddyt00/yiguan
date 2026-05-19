function escapeHtml(text) {
  return text.replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;').replace(/"/g,'&quot;')
}
function parseInline(text) {
  return text.replace(/\*\*(.+?)\*\*/g,'<strong>$1</strong>').replace(/\*(.+?)\*/g,'<em>$1</em>').replace(/`(.+?)`/g,'<code>$1</code>')
}
module.exports = {
  parse(text) {
    if (!text) return ''
    const lines = text.split('\n')
    const html = []
    let inCode = false, codeBuf = [], inList = false
    for (let i = 0; i < lines.length; i++) {
      const line = lines[i]
      if (/^```/.test(line.trim())) {
        if (inCode) { html.push('<pre><code>' + escapeHtml(codeBuf.join('\n')) + '</code></pre>'); codeBuf = []; inCode = false }
        else { inCode = true; codeBuf = [] }
        continue
      }
      if (inCode) { codeBuf.push(line); continue }
      if (line.trim() === '') { if (inList) { html.push('</ul>'); inList = false }; continue }
      if (/^#{1,6}\s/.test(line)) {
        const m = line.match(/^(#{1,6})\s+(.+)/)
        const lv = m[1].length
        if (inList) { html.push('</ul>'); inList = false }
        html.push('<h' + lv + '>' + parseInline(escapeHtml(m[2])) + '</h' + lv + '>')
        continue
      }
      if (/^\d+[.)]\s/.test(line)) {
        if (!inList) { html.push('<ol>'); inList = true }
        html.push('<li>' + parseInline(escapeHtml(line.replace(/^\d+[.)]\s/,''))) + '</li>')
        continue
      }
      if (/^[-*+]\s/.test(line)) {
        if (!inList) { html.push('<ul>'); inList = true }
        html.push('<li>' + parseInline(escapeHtml(line.replace(/^[-*+]\s/,''))) + '</li>')
        continue
      }
      if (inList) { html.push('</ul>'); inList = false }
      if (/^>\s/.test(line)) {
        html.push('<blockquote>' + parseInline(escapeHtml(line.replace(/^>\s?/,''))) + '</blockquote>')
        continue
      }
      html.push('<p>' + parseInline(escapeHtml(line)) + '</p>')
    }
    if (inCode) html.push('<pre><code>' + escapeHtml(codeBuf.join('\n')) + '</code></pre>')
    if (inList) html.push('</ul>')
    return html.join('\n')
  }
}
