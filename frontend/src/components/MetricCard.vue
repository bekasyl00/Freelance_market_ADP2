<script setup>
defineProps({
  label: String,
  value: [String, Number],
  detail: String,
  icon: Object,
});
</script>

<template>
  <article class="metric-card">
    <div class="metric-card__icon">
      <component :is="icon" :size="20" stroke-width="2" />
    </div>
    <div>
      <p>{{ label }}</p>
      <strong>{{ value }}</strong>
      <span>{{ detail }}</span>
    </div>
  </article>
</template>
