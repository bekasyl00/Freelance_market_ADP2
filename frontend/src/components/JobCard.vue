<script setup>
import { CalendarDays, CircleDollarSign, Send, ChevronRight } from 'lucide-vue-next';

defineProps({
  job: { type: Object, required: true },
  applied: Boolean,
  userRole: { type: String, default: 'freelancer' },
});

defineEmits(['apply', 'open']);
</script>

<template>
  <article class="job-card job-card--clickable" @click="$emit('open', job.id)">
    <div class="job-card__main">
      <div>
        <p class="eyebrow">{{ job.client }}</p>
        <h3>{{ job.title }}</h3>
      </div>
      <span class="status-pill" :class="`status-pill--${job.status}`">
        {{ job.status === 'open' ? $t('common.open') : job.status === 'completed' ? $t('common.completed') : $t('common.inProgress') }}
      </span>
    </div>

    <p class="job-card__description">{{ job.description }}</p>

    <div class="job-card__meta">
      <span><CircleDollarSign :size="16" /> {{ $t('jobs.budget') }}: {{ job.budget }} {{ $t('common.usd') }}</span>
      <span><CalendarDays :size="16" /> {{ $t('jobs.deadline') }}: {{ job.deadline || '—' }}</span>
    </div>

    <div class="skill-row">
      <span v-for="skill in job.skills" :key="skill">{{ skill }}</span>
    </div>

    <div class="job-card__footer">
      <!-- Only freelancers see Apply button -->
      <button
        v-if="userRole === 'freelancer' && job.status === 'open'"
        class="button button--primary"
        :disabled="applied"
        type="button"
        @click.stop="$emit('apply', job.id)"
      >
        <Send :size="16" />
        {{ applied ? $t('jobs.proposalSent') : $t('jobs.apply') }}
      </button>

      <button class="button button--ghost job-card__view-btn" type="button" @click.stop="$emit('open', job.id)">
        {{ $t('jobs.viewDetails') || 'View details' }}
        <ChevronRight :size="16" />
      </button>
    </div>
  </article>
</template>
