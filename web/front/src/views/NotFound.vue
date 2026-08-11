<template>
  <div class="nf">
    <div class="code">404</div>
    <h1>页面不存在</h1>
    <p>您访问的地址可能已失效，或被移动到了其他位置。请检查链接，或从首页重新进入。</p>
    <a class="op" href="/">返回首页</a>
  </div>
</template>

<style scoped>
.nf{
  height: 100%;  /* 占满 main（flex-1 自动填充视口剩余），footer 保留在底部，总高一屏内 */
  display:flex;
  flex-direction:column;
  align-items:center;
  justify-content:center;
  text-align:center;
}
.code{font-size:76px;font-weight:800;letter-spacing:-.03em;line-height:1;color:var(--gold);font-variant-numeric:tabular-nums}
h1{font-size:20px;font-weight:600;color:var(--text);margin-top:20px}
p{font-size:14px;color:var(--text-secondary);margin-top:10px;line-height:1.9;max-width:420px}
.op{display:inline-block;font-size:14px;margin-top:30px;padding:10px 28px;border-radius:6px;text-decoration:none;background:var(--gold);color:var(--bg-deep);font-weight:500}
.op:hover{filter:brightness(1.08)}
</style>
