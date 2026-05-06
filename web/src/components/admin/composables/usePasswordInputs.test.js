import { describe, it, expect } from 'vitest'
import { usePasswordInputs } from './usePasswordInputs.js'

describe('usePasswordInputs', () => {
  it('starts with empty values and not visible', () => {
    const pw = usePasswordInputs()
    expect(pw.password.value).toBe('')
    expect(pw.confirm.value).toBe('')
    expect(pw.visible.value).toBe(false)
  })

  it('toggleVisible flips the boolean', () => {
    const pw = usePasswordInputs()
    pw.toggleVisible()
    expect(pw.visible.value).toBe(true)
    pw.toggleVisible()
    expect(pw.visible.value).toBe(false)
  })

  it('isValid is false when empty', () => {
    const pw = usePasswordInputs()
    expect(pw.isValid.value).toBe(false)
  })

  it('isValid is false when confirm does not match', () => {
    const pw = usePasswordInputs()
    pw.password.value = 'ValidPass123Long'
    pw.confirm.value = 'DifferentPass1Long'
    expect(pw.isValid.value).toBe(false)
    expect(pw.errorMessage.value).toMatch(/match/i)
  })

  it('isValid is false when password fails complexity', () => {
    const pw = usePasswordInputs()
    pw.password.value = 'short'
    pw.confirm.value = 'short'
    expect(pw.isValid.value).toBe(false)
    expect(pw.errorMessage.value).toMatch(/12|character|upper|lower|digit/i)
  })

  it('isValid is true with matching valid password', () => {
    const pw = usePasswordInputs()
    pw.password.value = 'ValidPass123Long'
    pw.confirm.value = 'ValidPass123Long'
    expect(pw.isValid.value).toBe(true)
    expect(pw.errorMessage.value).toBe('')
  })

  it('reset clears all state', () => {
    const pw = usePasswordInputs()
    pw.password.value = 'ValidPass123Long'
    pw.confirm.value = 'ValidPass123Long'
    pw.visible.value = true
    pw.reset()
    expect(pw.password.value).toBe('')
    expect(pw.confirm.value).toBe('')
    expect(pw.visible.value).toBe(false)
  })
})
