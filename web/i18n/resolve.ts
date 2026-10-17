export type MessageParams = Record<string, string | number>

/** Walk dot paths on nested message objects. */
export function resolveMessage(tree: unknown, path: string): string | undefined {
  const parts = path.split('.')
  let cur: unknown = tree
  for (const part of parts) {
    if (cur === null || cur === undefined || typeof cur !== 'object') return undefined
    cur = (cur as Record<string, unknown>)[part]
  }
  return typeof cur === 'string' ? cur : undefined
}

export function formatMessage(template: string, params?: MessageParams): string {
  if (!params) return template
  return template.replace(/\{(\w+)\}/g, (_, key: string) => {
    const v = params[key]
    return v === undefined ? `{${key}}` : String(v)
  })
}
