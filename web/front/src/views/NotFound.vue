<template>
  <main class="nf">
    <div class="code">404</div>
    <h1>页面不存在</h1>
    <p>您访问的地址可能已失效，或被移动到了其他位置。请检查链接，或从首页重新进入。</p>
    <div class="ops">
      <a class="op" href="/">返回首页</a>
    </div>
  </main>
</template>

<style scoped>
.nf{min-height:100vh;display:flex;flex-direction:column;align-items:flex-start;justify-content:center;padding:0 8vw;background:var(--bg)}
.code{font-size:72px;font-weight:800;letter-spacing:-.03em;line-height:1;color:var(--text);font-variant-numeric:tabular-nums}
h1{font-size:18px;font-weight:600;color:var(--text);margin-top:22px}
p{font-size:14px;color:var(--text-secondary);margin-top:10px;line-height:1.9}
.ops{margin-top:34px}
.op{display:inline-block;font-size:14px;padding:9px 22px;border-radius:6px;text-decoration:none;background:var(--gold);color:var(--bg-deep)}
.op:hover{filter:brightness(1.08)}
</style>
