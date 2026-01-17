import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { shallowMount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import ImportView from './ImportView.vue'

// Mock axios
vi.mock('@/utils/axios', () => ({
  default: {
    post: vi.fn(),
  }
}))

import axios from '@/utils/axios'

describe('ImportView', () => {
  let wrapper

  const createWrapper = () => {
    const pinia = createPinia()
    setActivePinia(pinia)

    wrapper = shallowMount(ImportView, {
      global: {
        plugins: [pinia],
        stubs: {
          'v-container': { template: '<div><slot /></div>' },
          'v-card': { template: '<div><slot /></div>' },
          'v-btn': { template: '<button><slot /></button>' },
          'v-icon': { template: '<i></i>' },
          'v-alert': { template: '<div><slot /></div>' },
          'v-row': { template: '<div><slot /></div>' },
          'v-col': { template: '<div><slot /></div>' },
          'v-radio-group': { template: '<div><slot /></div>' },
          'v-radio': { template: '<input type="radio" />' },
          'v-checkbox': { template: '<input type="checkbox" />' },
          'v-chip': { template: '<span><slot /></span>' },
          'v-data-table': { template: '<table><slot /></table>' },
          'v-tooltip': { template: '<div><slot /></div>' },
          'v-bottom-navigation': { template: '<div><slot /></div>' }
        }
      }
    })

    return wrapper
  }

  beforeEach(() => {
    vi.clearAllMocks()
  })

  afterEach(() => {
    if (wrapper) {
      wrapper.unmount()
      wrapper = null
    }
  })

  // ==========================================================================
  // INITIALIZATION TESTS
  // ==========================================================================

  describe('Initialization', () => {
    it('initializes with wods as default entity', () => {
      createWrapper()

      const vm = wrapper.vm

      expect(vm.selectedEntity).toBe('wods')
    })

    it('initializes with no selected file', () => {
      createWrapper()

      const vm = wrapper.vm

      expect(vm.selectedFile).toBe(null)
    })

    it('initializes with no preview result', () => {
      createWrapper()

      const vm = wrapper.vm

      expect(vm.previewResult).toBe(null)
    })

    it('initializes with skip as default duplicate handling', () => {
      createWrapper()

      const vm = wrapper.vm

      expect(vm.duplicateHandling).toBe('skip')
    })

    it('initializes with skipDuplicates true', () => {
      createWrapper()

      const vm = wrapper.vm

      expect(vm.skipDuplicates).toBe(true)
    })
  })

  // ==========================================================================
  // ENTITY SELECTION TESTS
  // ==========================================================================

  describe('Entity Selection', () => {
    it('can select wods entity', () => {
      createWrapper()

      const vm = wrapper.vm
      vm.selectedEntity = 'wods'

      expect(vm.selectedEntity).toBe('wods')
      expect(vm.fileAccept).toBe('.csv')
    })

    it('can select movements entity', () => {
      createWrapper()

      const vm = wrapper.vm
      vm.selectedEntity = 'movements'

      expect(vm.selectedEntity).toBe('movements')
      expect(vm.fileAccept).toBe('.csv')
    })

    it('can select user_workouts entity', () => {
      createWrapper()

      const vm = wrapper.vm
      vm.selectedEntity = 'user_workouts'

      expect(vm.selectedEntity).toBe('user_workouts')
      expect(vm.fileAccept).toBe('.json')
    })

    it('can select wodify entity', () => {
      createWrapper()

      const vm = wrapper.vm
      vm.selectedEntity = 'wodify'

      expect(vm.selectedEntity).toBe('wodify')
      expect(vm.fileAccept).toBe('.csv')
    })
  })

  // ==========================================================================
  // COMPUTED PROPERTIES TESTS
  // ==========================================================================

  describe('Computed Properties', () => {
    it('fileAccept returns .csv for wods', () => {
      createWrapper()

      const vm = wrapper.vm
      vm.selectedEntity = 'wods'

      expect(vm.fileAccept).toBe('.csv')
    })

    it('fileAccept returns .json for user_workouts', () => {
      createWrapper()

      const vm = wrapper.vm
      vm.selectedEntity = 'user_workouts'

      expect(vm.fileAccept).toBe('.json')
    })

    it('fileTypeLabel returns CSV for wods', () => {
      createWrapper()

      const vm = wrapper.vm
      vm.selectedEntity = 'wods'

      expect(vm.fileTypeLabel).toBe('CSV')
    })

    it('fileTypeLabel returns JSON for user_workouts', () => {
      createWrapper()

      const vm = wrapper.vm
      vm.selectedEntity = 'user_workouts'

      expect(vm.fileTypeLabel).toBe('JSON')
    })

    it('isWodifyImport returns true for wodify', () => {
      createWrapper()

      const vm = wrapper.vm
      vm.selectedEntity = 'wodify'

      expect(vm.isWodifyImport).toBe(true)
    })

    it('isWodifyImport returns false for other entities', () => {
      createWrapper()

      const vm = wrapper.vm
      vm.selectedEntity = 'wods'

      expect(vm.isWodifyImport).toBe(false)
    })

    it('previewData returns empty array when no preview', () => {
      createWrapper()

      const vm = wrapper.vm

      expect(vm.previewData).toEqual([])
    })

    it('previewData returns rows from preview result', () => {
      createWrapper()

      const vm = wrapper.vm
      vm.previewResult = { rows: [{ name: 'Test' }] }

      expect(vm.previewData).toEqual([{ name: 'Test' }])
    })

    it('previewData returns workouts from preview result', () => {
      createWrapper()

      const vm = wrapper.vm
      vm.previewResult = { workouts: [{ workout_date: '2024-01-01' }] }

      expect(vm.previewData).toEqual([{ workout_date: '2024-01-01' }])
    })
  })

  // ==========================================================================
  // PREVIEW HEADERS TESTS
  // ==========================================================================

  describe('Preview Headers', () => {
    it('returns wods headers for wods entity', () => {
      createWrapper()

      const vm = wrapper.vm
      vm.selectedEntity = 'wods'

      const headers = vm.previewHeaders
      expect(headers.some(h => h.key === 'name')).toBe(true)
      expect(headers.some(h => h.key === 'type')).toBe(true)
      expect(headers.some(h => h.key === 'regime')).toBe(true)
    })

    it('returns movements headers for movements entity', () => {
      createWrapper()

      const vm = wrapper.vm
      vm.selectedEntity = 'movements'

      const headers = vm.previewHeaders
      expect(headers.some(h => h.key === 'name')).toBe(true)
      expect(headers.some(h => h.key === 'type')).toBe(true)
    })

    it('returns user_workouts headers for user_workouts entity', () => {
      createWrapper()

      const vm = wrapper.vm
      vm.selectedEntity = 'user_workouts'

      const headers = vm.previewHeaders
      expect(headers.some(h => h.key === 'workout_date')).toBe(true)
      expect(headers.some(h => h.key === 'wod_name')).toBe(true)
    })
  })

  // ==========================================================================
  // FILE HANDLING TESTS
  // ==========================================================================

  describe('File Handling', () => {
    it('handles file select', () => {
      createWrapper()

      const vm = wrapper.vm
      const mockFile = new File(['test'], 'test.csv', { type: 'text/csv' })

      vm.handleFileSelect({ target: { files: [mockFile] } })

      expect(vm.selectedFile).toEqual(mockFile)
    })

    it('handles file drop with valid csv', () => {
      createWrapper()

      const vm = wrapper.vm
      vm.selectedEntity = 'wods'
      const mockFile = new File(['test'], 'test.csv', { type: 'text/csv' })

      vm.handleDrop({
        dataTransfer: { files: [mockFile] }
      })

      expect(vm.selectedFile).toEqual(mockFile)
      expect(vm.dragOver).toBe(false)
    })

    it('handles file drop with valid json', () => {
      createWrapper()

      const vm = wrapper.vm
      vm.selectedEntity = 'user_workouts'
      const mockFile = new File(['{}'], 'test.json', { type: 'application/json' })

      vm.handleDrop({
        dataTransfer: { files: [mockFile] }
      })

      expect(vm.selectedFile).toEqual(mockFile)
    })

    it('rejects file drop with wrong extension', () => {
      createWrapper()

      const vm = wrapper.vm
      vm.selectedEntity = 'wods'
      const mockFile = new File(['{}'], 'test.json', { type: 'application/json' })

      vm.handleDrop({
        dataTransfer: { files: [mockFile] }
      })

      expect(vm.selectedFile).toBe(null)
      expect(vm.error).toContain('valid CSV file')
    })
  })

  // ==========================================================================
  // PREVIEW IMPORT TESTS
  // ==========================================================================

  describe('Preview Import', () => {
    it('does not preview without file', async () => {
      createWrapper()

      const vm = wrapper.vm
      vm.selectedFile = null

      await vm.previewImport()

      expect(axios.post).not.toHaveBeenCalled()
    })

    it('calls wods preview endpoint', async () => {
      axios.post.mockResolvedValue({
        data: { total_rows: 10, valid_rows: 8 }
      })
      createWrapper()

      const vm = wrapper.vm
      vm.selectedEntity = 'wods'
      vm.selectedFile = new File(['test'], 'test.csv')

      await vm.previewImport()
      await flushPromises()

      expect(axios.post).toHaveBeenCalledWith(
        '/api/import/wods/preview',
        expect.any(FormData),
        expect.objectContaining({ headers: { 'Content-Type': 'multipart/form-data' } })
      )
    })

    it('calls movements preview endpoint', async () => {
      axios.post.mockResolvedValue({
        data: { total_rows: 10, valid_rows: 8 }
      })
      createWrapper()

      const vm = wrapper.vm
      vm.selectedEntity = 'movements'
      vm.selectedFile = new File(['test'], 'test.csv')

      await vm.previewImport()
      await flushPromises()

      expect(axios.post).toHaveBeenCalledWith(
        '/api/import/movements/preview',
        expect.any(FormData),
        expect.any(Object)
      )
    })

    it('calls wodify preview endpoint', async () => {
      axios.post.mockResolvedValue({
        data: { total_rows: 10, valid_rows: 8 }
      })
      createWrapper()

      const vm = wrapper.vm
      vm.selectedEntity = 'wodify'
      vm.selectedFile = new File(['test'], 'test.csv')

      await vm.previewImport()
      await flushPromises()

      expect(axios.post).toHaveBeenCalledWith(
        '/api/import/wodify/preview',
        expect.any(FormData),
        expect.any(Object)
      )
    })

    it('calls user-workouts preview endpoint', async () => {
      axios.post.mockResolvedValue({
        data: { total_workouts: 10, valid_workouts: 8 }
      })
      createWrapper()

      const vm = wrapper.vm
      vm.selectedEntity = 'user_workouts'
      vm.selectedFile = new File(['{}'], 'test.json')

      await vm.previewImport()
      await flushPromises()

      expect(axios.post).toHaveBeenCalledWith(
        '/api/import/user-workouts/preview',
        expect.any(FormData),
        expect.any(Object)
      )
    })

    it('stores preview result', async () => {
      const mockResult = { total_rows: 10, valid_rows: 8, rows: [] }
      axios.post.mockResolvedValue({ data: mockResult })
      createWrapper()

      const vm = wrapper.vm
      vm.selectedFile = new File(['test'], 'test.csv')

      await vm.previewImport()
      await flushPromises()

      expect(vm.previewResult).toEqual(mockResult)
    })

    it('handles preview error', async () => {
      const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {})
      axios.post.mockRejectedValue({
        response: { data: { error: 'Invalid format' } }
      })
      createWrapper()

      const vm = wrapper.vm
      vm.selectedFile = new File(['test'], 'test.csv')

      await vm.previewImport()
      await flushPromises()

      expect(vm.error).toBe('Invalid format')

      consoleSpy.mockRestore()
    })

    it('sets uploading state during preview', async () => {
      let resolvePost
      axios.post.mockReturnValue(new Promise(resolve => { resolvePost = resolve }))

      createWrapper()

      const vm = wrapper.vm
      vm.selectedFile = new File(['test'], 'test.csv')

      const previewPromise = vm.previewImport()

      expect(vm.uploading).toBe(true)

      resolvePost({ data: {} })
      await previewPromise
      await flushPromises()

      expect(vm.uploading).toBe(false)
    })
  })

  // ==========================================================================
  // CONFIRM IMPORT TESTS
  // ==========================================================================

  describe('Confirm Import', () => {
    it('does not confirm without file', async () => {
      createWrapper()

      const vm = wrapper.vm
      vm.selectedFile = null

      await vm.confirmImport()

      expect(axios.post).not.toHaveBeenCalled()
    })

    it('does not confirm without preview', async () => {
      createWrapper()

      const vm = wrapper.vm
      vm.selectedFile = new File(['test'], 'test.csv')
      vm.previewResult = null

      await vm.confirmImport()

      expect(axios.post).not.toHaveBeenCalled()
    })

    it('calls wods confirm endpoint', async () => {
      axios.post.mockResolvedValue({
        data: { created_count: 5, updated_count: 2 }
      })
      createWrapper()

      const vm = wrapper.vm
      vm.selectedEntity = 'wods'
      vm.selectedFile = new File(['test'], 'test.csv')
      vm.previewResult = { valid_rows: 7 }
      vm.duplicateHandling = 'skip'

      await vm.confirmImport()
      await flushPromises()

      expect(axios.post).toHaveBeenCalledWith(
        '/api/import/wods/confirm',
        expect.any(FormData),
        expect.any(Object)
      )
    })

    it('calls movements confirm endpoint', async () => {
      axios.post.mockResolvedValue({
        data: { created_count: 5 }
      })
      createWrapper()

      const vm = wrapper.vm
      vm.selectedEntity = 'movements'
      vm.selectedFile = new File(['test'], 'test.csv')
      vm.previewResult = { valid_rows: 5 }

      await vm.confirmImport()
      await flushPromises()

      expect(axios.post).toHaveBeenCalledWith(
        '/api/import/movements/confirm',
        expect.any(FormData),
        expect.any(Object)
      )
    })

    it('stores import result and clears preview', async () => {
      const mockResult = { created_count: 5, updated_count: 2, skipped_count: 1 }
      axios.post.mockResolvedValue({ data: mockResult })
      createWrapper()

      const vm = wrapper.vm
      vm.selectedFile = new File(['test'], 'test.csv')
      vm.previewResult = { valid_rows: 8 }

      await vm.confirmImport()
      await flushPromises()

      expect(vm.importResult).toEqual(mockResult)
      expect(vm.previewResult).toBe(null)
      expect(vm.selectedFile).toBe(null)
    })

    it('handles confirm error', async () => {
      const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {})
      axios.post.mockRejectedValue({
        response: { data: { error: 'Import failed' } }
      })
      createWrapper()

      const vm = wrapper.vm
      vm.selectedFile = new File(['test'], 'test.csv')
      vm.previewResult = { valid_rows: 8 }

      await vm.confirmImport()
      await flushPromises()

      expect(vm.error).toBe('Import failed')

      consoleSpy.mockRestore()
    })

    it('sets confirming state during import', async () => {
      let resolvePost
      axios.post.mockReturnValue(new Promise(resolve => { resolvePost = resolve }))

      createWrapper()

      const vm = wrapper.vm
      vm.selectedFile = new File(['test'], 'test.csv')
      vm.previewResult = { valid_rows: 8 }

      const confirmPromise = vm.confirmImport()

      expect(vm.confirming).toBe(true)

      resolvePost({ data: {} })
      await confirmPromise
      await flushPromises()

      expect(vm.confirming).toBe(false)
    })
  })

  // ==========================================================================
  // RESET TESTS
  // ==========================================================================

  describe('Reset Import', () => {
    it('resets all state', () => {
      createWrapper()

      const vm = wrapper.vm
      vm.previewResult = { valid_rows: 8 }
      vm.importResult = { created_count: 5 }
      vm.selectedFile = new File(['test'], 'test.csv')
      vm.error = 'Some error'
      vm.duplicateHandling = 'update'

      vm.resetImport()

      expect(vm.previewResult).toBe(null)
      expect(vm.importResult).toBe(null)
      expect(vm.selectedFile).toBe(null)
      expect(vm.error).toBe(null)
      expect(vm.duplicateHandling).toBe('skip')
    })
  })

  // ==========================================================================
  // HELPER FUNCTION TESTS
  // ==========================================================================

  describe('Helper Functions', () => {
    it('getRowColor returns error for invalid rows', () => {
      createWrapper()

      const vm = wrapper.vm
      const result = vm.getRowColor({ is_valid: false })

      expect(result).toBe('error')
    })

    it('getRowColor returns warning for duplicate rows', () => {
      createWrapper()

      const vm = wrapper.vm
      const result = vm.getRowColor({ is_valid: true, is_duplicate: true })

      expect(result).toBe('warning')
    })

    it('getRowColor returns success for valid rows', () => {
      createWrapper()

      const vm = wrapper.vm
      const result = vm.getRowColor({ is_valid: true, is_duplicate: false })

      expect(result).toBe('success')
    })
  })
})
