import { describe, it, expect } from 'vitest'

// extracted logic from ToolsView for unit testing
function generateDefaultArgs(schema: object): string {
  const s = schema as { properties?: Record<string, { type: string }> }
  if (!s.properties) return '{}'
  const defaults: Record<string, unknown> = {}
  for (const [key, val] of Object.entries(s.properties)) {
    defaults[key] = val.type === 'string' ? '' : val.type === 'number' ? 0 : null
  }
  return JSON.stringify(defaults, null, 2)
}

describe('generateDefaultArgs', () => {
  it('returns empty object for schema with no properties', () => {
    expect(generateDefaultArgs({})).toBe('{}')
  })

  it('fills string fields with empty string', () => {
    const schema = { properties: { name: { type: 'string' } } }
    const result = JSON.parse(generateDefaultArgs(schema))
    expect(result.name).toBe('')
  })

  it('fills number fields with 0', () => {
    const schema = { properties: { count: { type: 'number' } } }
    const result = JSON.parse(generateDefaultArgs(schema))
    expect(result.count).toBe(0)
  })

  it('fills unknown types with null', () => {
    const schema = { properties: { flag: { type: 'boolean' } } }
    const result = JSON.parse(generateDefaultArgs(schema))
    expect(result.flag).toBeNull()
  })

  it('handles mixed schema properties', () => {
    const schema = {
      properties: {
        label: { type: 'string' },
        limit: { type: 'number' },
        active: { type: 'boolean' },
      },
    }
    const result = JSON.parse(generateDefaultArgs(schema))
    expect(result.label).toBe('')
    expect(result.limit).toBe(0)
    expect(result.active).toBeNull()
  })
})
