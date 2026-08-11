<template>
  <main class="nf">
    <div class="code">404</div>
    <h1>页面不存在</h1>
    <p>您访问的地址可能已失效，或被移动到了其他位置。请检查链接，或从管理后台重新进入。</p>
    <div class="ops">
      <a class="op" href="/admin">返回管理后台</a>
    </div>
  </main>
</template>

<style scoped>
.nf{min-height:100vh;display:flex;flex-direction:column;align-items:flex-start;justify-content:center;padding:0 8vw;background:var(--paper)}
.code{font-size:72px;font-weight:800;letter-spacing:-.03em;line-height:1;color:var(--ink);font-variant-numeric:tabular-nums}
h1{font-size:18px;font-weight:600;color:var(--ink);margin-top:22px}
p{font-size:14px;color:var(--muted);margin-top:10px;line-height:1.9}
.ops{margin-top:34px}
.op{display:inline-block;font-size:14px;padding:9px 22px;border-radius:6px;text-decoration:none;background:var(--ink);color:var(--paper)}
.op:hover{background:var(--ink-2)}
</style>
