<script setup>
/**
 * Shared footer action row used by every tab on the Admin User Edit screen.
 *
 * Props:
 *   draft  — the object returned by useUserDraft().
 *            Requires: isDirty (ComputedRef<boolean>), saving (Ref<boolean>),
 *            discard() (function), save() (function — called indirectly via
 *            click:save emit so parent tabs can intercept for pre-save dialogs).
 *
 * Emits:
 *   click:save — fired when the user clicks Save changes.
 *                Parent is responsible for calling draft.save() (or showing
 *                a confirmation dialog first, e.g. for email changes).
 *
 * The Discard button calls draft.discard() directly — no interception needed.
 */
defineProps({
  draft: { type: Object, required: true },
})

defineEmits(['click:save'])
</script>

<template>
  <v-divider class="my-4" />
  <div class="d-flex align-center">
    <v-fade-transition>
      <span
        v-if="draft.isDirty.value && !draft.saving.value"
        class="text-caption text-medium-emphasis"
      >
        · Unsaved changes
      </span>
    </v-fade-transition>
    <v-spacer />
    <v-btn
      variant="tonal"
      class="mr-2"
      :disabled="!draft.isDirty.value || draft.saving.value"
      @click="draft.discard()"
    >
      Discard
    </v-btn>
    <v-btn
      color="primary"
      :loading="draft.saving.value"
      :disabled="!draft.isDirty.value || draft.saving.value"
      @click="$emit('click:save')"
    >
      Save changes
    </v-btn>
  </div>
</template>
