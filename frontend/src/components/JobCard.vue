<script setup>
import { CalendarDays, CircleDollarSign, Send } from 'lucide-vue-next';

defineProps({
  job: {
    type: Object,
    required: true,
  },
  applied: Boolean,
});

defineEmits(['apply']);
</script>

<template>
  <article class="job-card">
    <div class="job-card__main">
      <div>
        <p class="eyebrow">{{ job.client }}</p>
        <h3>{{ job.title }}</h3>
      </div>
      <span class="status-pill" :class="`status-pill--${job.status}`">
        {{ job.status === 'open' ? $t('common.open') : $t('common.inProgress') }}
      </span>
    </div>

    <p class="job-card__description">{{ job.description }}</p>

    <div class="job-card__meta">
      <span><CircleDollarSign :size="16" /> {{ $t('jobs.budget') }}: {{ job.budget }} {{ $t('common.usd') }}</span>
      <span><CalendarDays :size="16" /> {{ $t('jobs.deadline') }}: {{ job.deadline }}</span>
    </div>

    <div class="skill-row">
      <span v-for="skill in job.skills" :key="skill">{{ skill }}</span>
    </div>

    <button class="button button--primary" :disabled="applied" type="button" @click="$emit('apply', job.id)">
      <Send :size="16" />
      {{ applied ? $t('jobs.proposalSent') : $t('jobs.apply') }}
    </button>
  </article>
</template>
