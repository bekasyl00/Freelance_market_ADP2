<script setup>
import { computed, onMounted, ref } from 'vue';
import { Plus, Search } from 'lucide-vue-next';
import JobCard from '../components/JobCard.vue';
import { marketplaceApi } from '../services/marketplace';

const jobs = ref([]);
const applied = ref(new Set());
const query = ref('');
const isPublishing = ref(false);
const newJob = ref({
  title: '',
  budget: 750,
  description: '',
  skills: 'Web Design, Branding, SEO',
});

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

async function applyToJob(jobId) {
  try {
    await marketplaceApi.applyToJob(jobId);
  } catch (error) {
    console.warn(error);
  }
  applied.value = new Set([...applied.value, jobId]);
}

async function publishJob() {
  if (!newJob.value.title.trim()) return;

  isPublishing.value = true;
  const payload = {
    title: newJob.value.title,
    budget: Number(newJob.value.budget) || 0,
    description: newJob.value.description,
    skills: newJob.value.skills.split(',').map((item) => item.trim()).filter(Boolean),
  };

  try {
    const created = await marketplaceApi.createJob(payload);
    jobs.value = [created, ...jobs.value];
  } catch (error) {
    console.warn(error);
    jobs.value = [
      {
        id: crypto.randomUUID(),
        title: payload.title,
        client: 'You',
        budget: payload.budget,
        deadline: '2026-06-20',
        status: 'open',
        skills: payload.skills,
        description: payload.description || 'New client request is ready to receive proposals.',
        proposals: 0,
      },
      ...jobs.value,
    ];
  } finally {
    isPublishing.value = false;
  }

  newJob.value = { title: '', budget: 750, description: '', skills: 'Web Design, Branding, SEO' };
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
      <label>
        {{ $t('jobs.skills') }}
        <input v-model="newJob.skills" type="text" />
      </label>
      <label class="job-form__wide">
        {{ $t('jobs.details') }}
        <input v-model="newJob.description" type="text" />
      </label>
      <button class="button button--primary" type="submit" :disabled="isPublishing">
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
