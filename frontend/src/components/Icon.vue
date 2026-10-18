<template>
  <svg
    xmlns="http://www.w3.org/2000/svg"
    :width="size"
    :height="size"
    viewBox="0 0 24 24"
    fill="none"
    :stroke="color"
    stroke-width="2"
    stroke-linecap="round"
    stroke-linejoin="round"
    :class="['lucide-icon', `lucide-${name}`]"
  >
    <!-- Sun icon -->
    <template v-if="name === 'sun'">
      <circle cx="12" cy="12" r="4"/>
      <path d="M12 2v2"/>
      <path d="M12 20v2"/>
      <path d="m4.93 4.93 1.41 1.41"/>
      <path d="m17.66 17.66 1.41 1.41"/>
      <path d="M2 12h2"/>
      <path d="M20 12h2"/>
      <path d="m6.34 17.66-1.41 1.41"/>
      <path d="m19.07 4.93-1.41 1.41"/>
    </template>

    <!-- Moon icon -->
    <template v-if="name === 'moon'">
      <path d="M20.985 12.486a9 9 0 1 1-9.473-9.472c.405-.022.617.46.402.803a6 6 0 0 0 8.268 8.268c.344-.215.825-.004.803.401"/>
    </template>

    <!-- Refresh icon -->
    <template v-if="name === 'refresh'">
      <path d="M21 12a9 9 0 0 0-15.5-6.36L3 8"/>
      <path d="M3 3v5h5"/>
      <path d="M3 12a9 9 0 0 0 15.5 6.36L21 16"/>
      <path d="M16 16h5v5"/>
    </template>

    <!-- List Tree icon -->
    <template v-if="name === 'list-tree'">
      <path d="M21 12h-8"/>
      <path d="M21 6h-8"/>
      <path d="M21 18h-8"/>
      <path d="M3 6v4c0 1.1.9 2 2 2h3"/>
      <path d="M3 10v6c0 1.1.9 2 2 2h3"/>
      <circle cx="3" cy="6" r="1"/>
      <circle cx="3" cy="12" r="1"/>
      <circle cx="3" cy="18" r="1"/>
    </template>

    <!-- Arrow Left / Back icon -->
    <template v-if="name === 'arrow-left'">
      <path d="m12 19-7-7 7-7"/>
      <path d="M19 12H5"/>
    </template>
  </svg>
</template>

<script setup lang="ts">
interface Props {
  name: 'refresh' | 'sun' | 'moon' | 'list-tree' | 'arrow-left'
  size?: number | string
  color?: string
}

withDefaults(defineProps<Props>(), {
  size: 20,
  color: 'currentColor'
})
</script>

<style scoped>
.lucide-icon {
  transition: transform 0.2s ease;
}
</style>
