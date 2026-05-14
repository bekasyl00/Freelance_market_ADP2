<script setup>
import { computed, inject, onMounted, ref } from 'vue';
import { Plus, Search } from 'lucide-vue-next';
import JobCard from '../components/JobCard.vue';
import { marketplaceApi } from '../services/marketplace';

const navigateTo = inject('navigateTo');
const jobs = ref([]);
const applied = ref(new Set());
const query = ref('');

const userRole = computed(() => localStorage.getItem('fm_user_role') || 'freelancer');
const isClient = computed(() => userRole.value === 'client');

const filteredJobs = computed(() => {
  const term = query.value.trim().toLowerCase();
  if (!term) return jobs.value;
  return jobs.value.filter((job) => {
    return [job.title, job.client, job.description, ...job.skills].join(' ').toLowerCase().includes(term);
  });
});

onMounted(async () => {
  jobs.value = await marketplaceApi.getJobs();
  const userId = localStorage.getItem('fm_user_id');
  if (userId) {
    const appliedIds = jobs.value
      .filter(j => j.freelancers && j.freelancers.includes(userId))
      .map(j => j.id);
    applied.value = new Set(appliedIds);
  }
});

async function applyToJob(jobId) {
  try {
    await marketplaceApi.applyToJob(jobId);
    applied.value = new Set([...applied.value, jobId]);
    const job = jobs.value.find(j => j.id === jobId);
    if (job && job.clientId) {
      navigateTo('chat:' + job.clientId + '|' + jobId);
    }
  } catch (error) {
    console.warn(error);
  }
}

function openJob(jobId) {
  navigateTo('jobDetail:' + jobId);
}
</script>

<template>
  <section class="view-stack">
    <div class="section-heading">
      <div>
        <p class="eyebrow">{{ $t('jobs.eyebrow') }}</p>
        <h1>{{ $t('jobs.title') }}</h1>
      </div>
      <p>{{ $t('jobs.subtitle') }}</p>
    </div>

    <div class="toolbar">
      <label class="search-field">
        <Search :size="18" />
        <input v-model="query" :placeholder="$t('app.search')" type="search" />
      </label>
      <button v-if="isClient" class="button button--primary" @click="navigateTo('createJob')">
        <Plus :size="16" />
        {{ $t('jobs.create') }}
      </button>
    </div>

    <div class="jobs-grid" v-if="filteredJobs.length">
      <JobCard
        v-for="job in filteredJobs"
        :key="job.id"
        :job="job"
        :applied="applied.has(job.id)"
        :user-role="userRole"
        @apply="applyToJob"
        @open="openJob"
      />
    </div>
    <p v-else class="empty-state">{{ $t('jobs.listEmpty') }}</p>
  </section>
</template>
