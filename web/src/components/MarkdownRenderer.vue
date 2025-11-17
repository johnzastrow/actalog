<template>
  <div class="markdown-content" v-html="renderedHtml"></div>
</template>

<script setup>
import { computed } from 'vue'
import { marked } from 'marked'

const props = defineProps({
  content: {
    type: String,
    default: ''
  }
})

// Configure marked options for security and formatting
marked.setOptions({
  breaks: true, // Convert \n to <br>
  gfm: true, // GitHub Flavored Markdown
  headerIds: false, // Disable header IDs
  mangle: false // Don't mangle email addresses
})

// Render markdown to HTML
const renderedHtml = computed(() => {
  if (!props.content) return ''

  try {
    return marked.parse(props.content)
  } catch (error) {
    console.error('Failed to parse markdown:', error)
    return props.content // Fallback to plain text
  }
})
</script>

<style scoped>
.markdown-content {
  line-height: 1.6;
  word-wrap: break-word;
}

/* Headings */
.markdown-content :deep(h1) {
  font-size: 1.5em;
  font-weight: bold;
  margin: 0.5em 0;
  color: #1a1a1a;
}

.markdown-content :deep(h2) {
  font-size: 1.3em;
  font-weight: bold;
  margin: 0.5em 0;
  color: #1a1a1a;
}

.markdown-content :deep(h3) {
  font-size: 1.1em;
  font-weight: bold;
  margin: 0.5em 0;
  color: #1a1a1a;
}

/* Paragraphs */
.markdown-content :deep(p) {
  margin: 0.5em 0;
}

/* Lists */
.markdown-content :deep(ul),
.markdown-content :deep(ol) {
  margin: 0.5em 0;
  padding-left: 1.5em;
}

.markdown-content :deep(li) {
  margin: 0.25em 0;
}

/* Code */
.markdown-content :deep(code) {
  background-color: #f5f7fa;
  padding: 0.2em 0.4em;
  border-radius: 3px;
  font-family: 'Courier New', monospace;
  font-size: 0.9em;
  color: #e91e63;
}

.markdown-content :deep(pre) {
  background-color: #f5f7fa;
  padding: 1em;
  border-radius: 5px;
  overflow-x: auto;
  margin: 0.5em 0;
}

.markdown-content :deep(pre code) {
  background-color: transparent;
  padding: 0;
  color: #1a1a1a;
}

/* Links */
.markdown-content :deep(a) {
  color: #00bcd4;
  text-decoration: none;
}

.markdown-content :deep(a:hover) {
  text-decoration: underline;
}

/* Emphasis */
.markdown-content :deep(strong) {
  font-weight: bold;
}

.markdown-content :deep(em) {
  font-style: italic;
}

/* Blockquotes */
.markdown-content :deep(blockquote) {
  border-left: 4px solid #00bcd4;
  padding-left: 1em;
  margin: 0.5em 0;
  color: #666;
  font-style: italic;
}

/* Horizontal rules */
.markdown-content :deep(hr) {
  border: none;
  border-top: 1px solid #ccc;
  margin: 1em 0;
}

/* Tables */
.markdown-content :deep(table) {
  border-collapse: collapse;
  width: 100%;
  margin: 0.5em 0;
}

.markdown-content :deep(th),
.markdown-content :deep(td) {
  border: 1px solid #ccc;
  padding: 0.5em;
  text-align: left;
}

.markdown-content :deep(th) {
  background-color: #f5f7fa;
  font-weight: bold;
}
</style>
