export const PROVIDERS = [
  {
    key: 'openrouter', label: 'OpenRouter', baseURL: 'https://openrouter.ai/api/v1',
    models: ['openai/gpt-4o', 'anthropic/claude-sonnet-4', 'google/gemini-2.5-flash', 'meta-llama/llama-3.1-405b'],
  },
  {
    key: 'vercel', label: 'Vercel AI Gateway', baseURL: 'https://api.vercel.ai/v1',
    models: ['gpt-4o', 'claude-sonnet-4', 'gemini-2.5-pro'],
  },
  {
    key: 'anthropic', label: 'Anthropic', baseURL: 'https://api.anthropic.com/v1',
    models: ['claude-sonnet-4-20250514', 'claude-haiku-3-5', 'claude-opus-4-20250514'],
  },
  {
    key: 'mimo', label: 'Xiaomi MiMo', baseURL: 'https://api.mimo.xiaomi.com/v1',
    models: ['MiMo-V2.5-Pro', 'MiMo-V2.5-Omni', 'MiMo-V2.5-Flash'],
  },
  {
    key: 'nvidia', label: 'NVIDIA NIM', baseURL: 'https://integrate.api.nvidia.com/v1',
    models: ['nemotron-4', 'llama-3.1-nemotron-70b'],
  },
  {
    key: 'huggingface', label: 'Hugging Face', baseURL: 'https://api-inference.huggingface.co/v1',
    models: ['meta-llama/Llama-3.1-70B', 'mistralai/Mixtral-8x22B', 'Qwen/Qwen2.5-72B'],
  },
  {
    key: 'google', label: 'Google AI Studio', baseURL: 'https://generativelanguage.googleapis.com/v1beta/openai',
    models: ['gemini-2.5-flash', 'gemini-2.5-pro', 'gemini-2.0-flash'],
  },
  {
    key: 'deepseek', label: 'DeepSeek', baseURL: 'https://api.deepseek.com/v1',
    models: ['deepseek-chat', 'deepseek-reasoner'],
  },
  {
    key: 'xai', label: 'xAI (Grok)', baseURL: 'https://api.x.ai/v1',
    models: ['grok-3', 'grok-3-mini'],
  },
  {
    key: 'zhipu', label: '智谱 AI (Z.AI/GLM)', baseURL: 'https://open.bigmodel.cn/api/paas/v4',
    models: ['glm-4-plus', 'glm-4-flash', 'glm-4'],
  },
  {
    key: 'kimi', label: 'Kimi / Moonshot', baseURL: 'https://api.moonshot.cn/v1',
    models: ['moonshot-v1-8k', 'moonshot-v1-32k', 'moonshot-v1-128k'],
  },
  {
    key: 'stepfun', label: 'StepFun', baseURL: 'https://api.stepfun.com/v1',
    models: ['step-2-16k', 'step-1-flash'],
  },
  {
    key: 'minimax', label: 'MiniMax (国际)', baseURL: 'https://api.minimax.chat/v1',
    models: ['abab7-chat', 'abab6.5s-chat'],
  },
  {
    key: 'minimax-cn', label: 'MiniMax (国内)', baseURL: 'https://api.minimaxi.com/v1',
    models: ['abab7-chat', 'abab6.5s-chat'],
  },
  {
    key: 'dashscope', label: '阿里云 DashScope', baseURL: 'https://dashscope.aliyuncs.com/compatible-mode/v1',
    models: ['qwen-plus', 'qwen-max', 'qwen-turbo', 'qwen3.6-plus'],
  },
  {
    key: 'ollama-cloud', label: 'Ollama Cloud', baseURL: 'https://api.ollama.com/v1',
    models: ['llama3.2', 'mistral', 'qwen2.5'],
  },
  {
    key: 'arcee', label: 'Arcee AI', baseURL: 'https://api.arcee.ai/v1',
    models: ['trinity', 'trinity-mini'],
  },
  {
    key: 'gmi', label: 'GMI Cloud', baseURL: 'https://api.gmicloud.com/v1',
    models: [],
  },
  {
    key: 'kilo', label: 'Kilo Code', baseURL: 'https://api.kilocode.ai/v1',
    models: [],
  },
  {
    key: 'opencode-zen', label: 'OpenCode Zen', baseURL: 'https://api.opencode.ai/zen/v1',
    models: [],
  },
  {
    key: 'opencode-go', label: 'OpenCode Go', baseURL: 'https://api.opencode.ai/go/v1',
    models: [],
  },
  {
    key: 'bedrock', label: 'AWS Bedrock', baseURL: '',
    models: ['anthropic.claude-sonnet-4', 'amazon.nova-pro'],
  },
  {
    key: 'azure', label: 'Azure Foundry', baseURL: '',
    models: [],
  },
  {
    key: 'qwen-code', label: 'Qwen Coding Plan', baseURL: 'https://coding.dashscope.aliyuncs.com/v1',
    models: [
      'qwen3.6-plus', 'qwen3.5-plus', 'qwen3-max-2026-01-23',
      'qwen3-coder-next', 'qwen3-coder-plus',
      'glm-5', 'glm-4.7',
      'kimi-k2.5',
      'MiniMax-M2.5',
    ],
  },
  {
    key: 'custom', label: '自定义', baseURL: '',
    models: [],
  },
]

export function findProvider(key) {
  return PROVIDERS.find(p => p.key === key) || PROVIDERS.find(p => p.key === 'custom')
}

export function providerLabel(key) {
  const p = findProvider(key)
  return p ? p.label : key
}
