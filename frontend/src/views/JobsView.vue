<script setup>
import { computed, onMounted, ref } from 'vue';
import { Plus, Search } from 'lucide-vue-next';
import JobCard from '../components/JobCard.vue';
import { marketplaceApi } from '../services/marketplace';

const jobs = ref([]);
const applied = ref(new Set());
const query = ref('');
const newJob = ref({ title: '', budget: 750 });

const filteredJobs = computed(() => {
  const term = query.value.trim().toLowerCase();
  if (!term) return jobs.value;
  return jobs.value.filter((job) => {
    return [job.title, job.client, job.description, ...job.skills].join(' ').toLowerCase().includes(term);
  });
});

onMounted(async () => {
  jobs.value = await marketplaceApi.getJobs();
});

function applyToJob(jobId) {
  applied.value = new Set([...applied.value, jobId]);
}

function publishJob() {
  if (!newJob.value.title.trim()) return;

  jobs.value = [
    {
      id: crypto.randomUUID(),
      title: newJob.value.title,
      client: 'You',
      budget: Number(newJob.value.budget) || 0,
      deadline: '2026-06-20',
      status: 'open',
      skills: ['Vue', 'Go', 'Postgres'],
      description: 'New client request is ready to receive proposals.',
      proposals: 0,
    },
    ...jobs.value,
  ];

  newJob.value = { title: '', budget: 750 };
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
    </div>

    <form class="job-form" @submit.prevent="publishJob">
      <label>
        {{ $t('jobs.newTitle') }}
        <input v-model="newJob.title" type="text" required />
      </label>
      <label>
        {{ $t('jobs.newBudget') }}
        <input v-model="newJob.budget" min="100" step="50" type="number" />
      </label>
      <button class="button button--primary" type="submit">
        <Plus :size="16" />
        {{ $t('jobs.publish') }}
      </button>
    </form>

    <div class="jobs-grid" v-if="filteredJobs.length">
      <JobCard
        v-for="job in filteredJobs"
        :key="job.id"
        :job="job"
        :applied="applied.has(job.id)"
        @apply="applyToJob"
      />
    </div>
    <p v-else class="empty-state">{{ $t('jobs.listEmpty') }}</p>
  </section>
</template>
