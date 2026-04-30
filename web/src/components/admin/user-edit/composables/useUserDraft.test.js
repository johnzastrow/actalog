import { describe, it, expect, vi, beforeEach } from 'vitest'
import { useUserDraft } from './useUserDraft'

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

/**
 * Build a draft for the standard profile fields used across all tests.
 */
function makeProfileDraft({ fetchOriginal, saveFields } = {}) {
  return useUserDraft({
    fetchOriginal: fetchOriginal ?? vi.fn(),
    saveFields: saveFields ?? vi.fn(),
    fields: ['name', 'email', 'birthday', 'email_verified'],
  })
}

const ORIGINAL_USER = {
  id: 42,
  name: 'Alice',
  email: 'alice@example.com',
  birthday: '1990-01-15',
  email_verified: true,
  updated_at: '2026-04-01T10:00:00Z',
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

describe('useUserDraft', () => {
  describe('starts clean after fetchOriginal resolves', () => {
    it('isDirty is false after load() completes', async () => {
      const fetchOriginal = vi.fn().mockResolvedValue({ data: ORIGINAL_USER })
      const draft = makeProfileDraft({ fetchOriginal })

      await draft.load()

      expect(draft.isDirty.value).toBe(false)
    })

    it('modified set is empty after load()', async () => {
      const fetchOriginal = vi.fn().mockResolvedValue({ data: ORIGINAL_USER })
      const draft = makeProfileDraft({ fetchOriginal })

      await draft.load()

      expect(draft.modified.value.size).toBe(0)
    })

    it('working values match original after load()', async () => {
      const fetchOriginal = vi.fn().mockResolvedValue({ data: ORIGINAL_USER })
      const draft = makeProfileDraft({ fetchOriginal })

      await draft.load()

      expect(draft.working.name).toBe(ORIGINAL_USER.name)
      expect(draft.working.email).toBe(ORIGINAL_USER.email)
    })
  })

  describe('load() surfaces errors', () => {
    it('sets error.value and re-throws when fetchOriginal rejects', async () => {
      const fetchError = new Error('Network failure')
      const fetchOriginal = vi.fn().mockRejectedValue(fetchError)
      const draft = makeProfileDraft({ fetchOriginal })

      await expect(draft.load()).rejects.toBe(fetchError)

      expect(draft.error.value).toBe(fetchError)
    })
  })

  describe('detects modified fields when working differs from original', () => {
    it('adds field to modified set when working value changes', async () => {
      const fetchOriginal = vi.fn().mockResolvedValue({ data: ORIGINAL_USER })
      const draft = makeProfileDraft({ fetchOriginal })
      await draft.load()

      draft.working.name = 'NewName'

      expect(draft.modified.value.has('name')).toBe(true)
    })

    it('isDirty becomes true when any field is changed', async () => {
      const fetchOriginal = vi.fn().mockResolvedValue({ data: ORIGINAL_USER })
      const draft = makeProfileDraft({ fetchOriginal })
      await draft.load()

      draft.working.name = 'NewName'

      expect(draft.isDirty.value).toBe(true)
    })

    it('unmodified fields are not in modified set', async () => {
      const fetchOriginal = vi.fn().mockResolvedValue({ data: ORIGINAL_USER })
      const draft = makeProfileDraft({ fetchOriginal })
      await draft.load()

      draft.working.name = 'NewName'

      expect(draft.modified.value.has('email')).toBe(false)
      expect(draft.modified.value.has('birthday')).toBe(false)
    })
  })

  describe("save() sends only modified fields plus updated_at", () => {
    it('payload contains only the changed field and updated_at', async () => {
      const fetchOriginal = vi.fn().mockResolvedValue({ data: ORIGINAL_USER })
      const savedUser = { ...ORIGINAL_USER, name: 'NewName' }
      const saveFields = vi.fn().mockResolvedValue({ data: savedUser })
      const draft = makeProfileDraft({ fetchOriginal, saveFields })

      await draft.load()
      draft.working.name = 'NewName'
      await draft.save()

      expect(saveFields).toHaveBeenCalledOnce()
      const [payload] = saveFields.mock.calls[0]
      expect(payload).toHaveProperty('name', 'NewName')
      expect(payload).toHaveProperty('updated_at', ORIGINAL_USER.updated_at)
      // Must NOT include unmodified fields
      expect(payload).not.toHaveProperty('email')
      expect(payload).not.toHaveProperty('birthday')
      expect(payload).not.toHaveProperty('email_verified')
    })

    it('does not call saveFields when nothing is dirty', async () => {
      const fetchOriginal = vi.fn().mockResolvedValue({ data: ORIGINAL_USER })
      const saveFields = vi.fn()
      const draft = makeProfileDraft({ fetchOriginal, saveFields })

      await draft.load()
      await draft.save()

      expect(saveFields).not.toHaveBeenCalled()
    })

    it('concurrent save() calls only invoke saveFields once', async () => {
      const fetchOriginal = vi.fn().mockResolvedValue({ data: ORIGINAL_USER })
      let resolveFirst
      const deferred = new Promise((resolve) => { resolveFirst = resolve })
      const saveFields = vi.fn().mockReturnValue(deferred)
      const draft = makeProfileDraft({ fetchOriginal, saveFields })

      await draft.load()
      draft.working.name = 'NewName'

      // First call starts the in-flight save (saving becomes true, does not await yet).
      const firstSave = draft.save()
      // Second call should be a no-op because saving.value is already true.
      const secondSave = draft.save()
      await secondSave  // resolves immediately

      // Resolve the deferred so the first save can complete.
      resolveFirst({ data: { ...ORIGINAL_USER, name: 'NewName' } })
      await firstSave

      expect(saveFields).toHaveBeenCalledOnce()
    })
  })

  describe("save() refreshes original and clears dirty on success", () => {
    it('isDirty becomes false after a successful save', async () => {
      const fetchOriginal = vi.fn().mockResolvedValue({ data: ORIGINAL_USER })
      const savedUser = { ...ORIGINAL_USER, name: 'NewName', updated_at: '2026-04-02T10:00:00Z' }
      const saveFields = vi.fn().mockResolvedValue({ data: savedUser })
      const draft = makeProfileDraft({ fetchOriginal, saveFields })

      await draft.load()
      draft.working.name = 'NewName'
      await draft.save()

      expect(draft.isDirty.value).toBe(false)
    })

    it('original reflects server response after save', async () => {
      const fetchOriginal = vi.fn().mockResolvedValue({ data: ORIGINAL_USER })
      const savedUser = { ...ORIGINAL_USER, name: 'NewName', updated_at: '2026-04-02T10:00:00Z' }
      const saveFields = vi.fn().mockResolvedValue({ data: savedUser })
      const draft = makeProfileDraft({ fetchOriginal, saveFields })

      await draft.load()
      draft.working.name = 'NewName'
      await draft.save()

      expect(draft.original.value.name).toBe('NewName')
      expect(draft.original.value.updated_at).toBe('2026-04-02T10:00:00Z')
    })

    it('saving flag is false after save resolves', async () => {
      const fetchOriginal = vi.fn().mockResolvedValue({ data: ORIGINAL_USER })
      const saveFields = vi.fn().mockResolvedValue({ data: ORIGINAL_USER })
      const draft = makeProfileDraft({ fetchOriginal, saveFields })

      await draft.load()
      draft.working.name = 'NewName'
      await draft.save()

      expect(draft.saving.value).toBe(false)
    })
  })

  describe("save() surfaces 409 as conflict", () => {
    it('sets conflict.value to true on 409 response', async () => {
      const fetchOriginal = vi.fn().mockResolvedValue({ data: ORIGINAL_USER })
      const conflictError = Object.assign(new Error('Conflict'), {
        response: { status: 409, data: { error: 'stale data' } },
      })
      const saveFields = vi.fn().mockRejectedValue(conflictError)
      const draft = makeProfileDraft({ fetchOriginal, saveFields })

      await draft.load()
      draft.working.name = 'NewName'

      await expect(draft.save()).rejects.toThrow('Conflict')

      expect(draft.conflict.value).toBe(true)
    })

    it('re-throws the error after setting conflict', async () => {
      const fetchOriginal = vi.fn().mockResolvedValue({ data: ORIGINAL_USER })
      const conflictError = Object.assign(new Error('Conflict'), {
        response: { status: 409 },
      })
      const saveFields = vi.fn().mockRejectedValue(conflictError)
      const draft = makeProfileDraft({ fetchOriginal, saveFields })

      await draft.load()
      draft.working.name = 'NewName'

      await expect(draft.save()).rejects.toBe(conflictError)
    })

    it('does not set conflict for non-409 errors', async () => {
      const fetchOriginal = vi.fn().mockResolvedValue({ data: ORIGINAL_USER })
      const serverError = Object.assign(new Error('Internal Server Error'), {
        response: { status: 500 },
      })
      const saveFields = vi.fn().mockRejectedValue(serverError)
      const draft = makeProfileDraft({ fetchOriginal, saveFields })

      await draft.load()
      draft.working.name = 'NewName'

      await expect(draft.save()).rejects.toThrow()

      expect(draft.conflict.value).toBe(false)
    })
  })

  describe("discard() resets working to original", () => {
    it('working values match original after discard', async () => {
      const fetchOriginal = vi.fn().mockResolvedValue({ data: ORIGINAL_USER })
      const draft = makeProfileDraft({ fetchOriginal })

      await draft.load()
      draft.working.name = 'NewName'
      draft.working.email = 'changed@example.com'
      draft.discard()

      expect(draft.working.name).toBe(ORIGINAL_USER.name)
      expect(draft.working.email).toBe(ORIGINAL_USER.email)
    })

    it('isDirty is false after discard', async () => {
      const fetchOriginal = vi.fn().mockResolvedValue({ data: ORIGINAL_USER })
      const draft = makeProfileDraft({ fetchOriginal })

      await draft.load()
      draft.working.name = 'NewName'
      draft.discard()

      expect(draft.isDirty.value).toBe(false)
    })

    it('clears conflict flag on discard', async () => {
      const fetchOriginal = vi.fn().mockResolvedValue({ data: ORIGINAL_USER })
      const conflictError = Object.assign(new Error('Conflict'), {
        response: { status: 409 },
      })
      const saveFields = vi.fn().mockRejectedValue(conflictError)
      const draft = makeProfileDraft({ fetchOriginal, saveFields })

      await draft.load()
      draft.working.name = 'NewName'
      await draft.save().catch(() => {})

      draft.discard()

      expect(draft.conflict.value).toBe(false)
    })
  })
})
