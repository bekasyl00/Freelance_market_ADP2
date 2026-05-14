<script setup>
import { ref, inject, onMounted, computed } from 'vue';
import { ArrowLeft, CalendarDays, CircleDollarSign, Send, Users, User, CheckCircle, MessageSquare } from 'lucide-vue-next';
import { marketplaceApi } from '../services/marketplace';

const props = defineProps({ jobId: { type: String, default: '' } });
const navigateTo = inject('navigateTo');

const job = ref(null);
const applying = ref(false);
const applied = ref(false);
const error = ref('');

const userRole = computed(() => localStorage.getItem('fm_user_role') || 'freelancer');

onMounted(async () => {
  try {
    job.value = await marketplaceApi.getJobDetail(props.jobId);
  } catch {
    // fallback: try from cached jobs list
    try {
      const all = await marketplaceApi.getJobs();
      job.value = all.find(j => j.id === props.jobId) || null;
    } catch {
      job.value = null;
    }
  }
  const userId = localStorage.getItem('fm_user_id');
  if (job.value && job.value.freelancers && userId) {
    if (job.value.freelancers.includes(userId)) {
      applied.value = true;
    }
  }
});

async function applyToJob() {
  if (!job.value) return;
  applying.value = true;
  try {
    await marketplaceApi.applyToJob(job.value.id);
    applied.value = true;
    // User wants a chat to be created automatically after applying
    if (job.value.clientId) {
      navigateTo('chat:' + job.value.clientId + '|' + job.value.id);
    }
  } catch (e) {
    error.value = e.message;
  } finally {
    applying.value = false;
  }
}

async function completeJob() {
  if (!job.value) return;
  try {
    await marketplaceApi.completeJob(job.value.id);
    job.value.status = 'completed';
  } catch (e) {
    error.value = e.message;
  }
}

function statusLabel(status) {
  const map = { open: 'Open', inProgress: 'In Progress', completed: 'Completed' };
  return map[status] || status;
}
</script>

<template>
  <section class="view-stack">
    <div class="section-heading">
      <div>
        <button class="button button--ghost" type="button" @click="navigateTo('jobs')" style="margin-bottom:12px">
          <ArrowLeft :size="16" />
          {{ $t('jobs.backToJobs') || 'Back to jobs' }}
        </button>
      </div>
    </div>

    <div v-if="!job" class="chat-empty-state">
      <p>{{ $t('common.loading') || 'Loading...' }}</p>
    </div>

    <div v-else class="job-detail">
      <div class="job-detail__header">
        <div>
          <p class="eyebrow">{{ job.client }}</p>
          <h1 class="job-detail__title">{{ job.title }}</h1>
        </div>
        <span class="status-pill" :class="`status-pill--${job.status}`">
          {{ statusLabel(job.status) }}
        </span>
      </div>

      <div class="job-detail__meta">
        <div class="job-detail__meta-item">
          <CircleDollarSign :size="20" />
          <div>
            <span class="job-detail__meta-label">{{ $t('jobs.budget') }}</span>
            <strong>${{ job.budget }}</strong>
          </div>
        </div>
        <div class="job-detail__meta-item">
          <CalendarDays :size="20" />
          <div>
            <span class="job-detail__meta-label">{{ $t('jobs.deadline') }}</span>
            <strong>{{ job.deadline || '—' }}</strong>
          </div>
        </div>
        <div class="job-detail__meta-item">
          <Users :size="20" />
          <div>
            <span class="job-detail__meta-label">{{ $t('jobs.proposals') || 'Proposals' }}</span>
            <strong>{{ job.proposals || 0 }}</strong>
          </div>
        </div>
      </div>

      <div class="job-detail__body">
        <h2>{{ $t('jobs.description') || 'Description' }}</h2>
        <p>{{ job.description }}</p>
      </div>

      <div class="job-detail__skills">
        <h3>{{ $t('jobs.skills') }}</h3>
        <div class="skill-row skill-row--large">
          <span v-for="skill in job.skills" :key="skill">{{ skill }}</span>
        </div>
      </div>

      <div v-if="error" class="auth-error" style="margin-top:16px">{{ error }}</div>

      <div class="job-detail__actions">
        <!-- Freelancer can apply -->
        <button
          v-if="userRole === 'freelancer' && job.status === 'open'"
          class="button button--primary"
          :disabled="applied || applying"
          @click="applyToJob"
        >
          <Send :size="16" />
          {{ applied ? ($t('jobs.proposalSent') || 'Proposal sent') : ($t('jobs.apply') || 'Apply') }}
        </button>

        <button
          v-if="userRole === 'freelancer' && job.clientId"
          class="button button--ghost"
          @click="navigateTo('chat:' + job.clientId)"
        >
          <MessageSquare :size="16" />
          {{ $t('jobs.messageClient') || 'Message Client' }}
        </button>

        <!-- Client can complete -->
        <button
          v-if="userRole === 'client' && job.status === 'inProgress'"
          class="button button--primary"
          @click="completeJob"
        >
          <CheckCircle :size="16" />
          {{ $t('jobs.markComplete') || 'Mark as Complete' }}
        </button>
      </div>

      <!-- Client can message the freelancers who applied -->
      <div v-if="userRole === 'client' && job.applicants && job.applicants.length > 0" class="job-detail__applicants" style="margin-top: 2rem; border-top: 1px solid var(--border); padding-top: 2rem;">
        <h3>{{ $t('jobs.applicants') || 'Откликнувшиеся фрилансеры' }}</h3>
        <div class="applicants-list" style="display: flex; flex-direction: column; gap: 1rem; margin-top: 1rem;">
          <div v-for="app in job.applicants" :key="app.id" class="applicant-card" style="display: flex; align-items: center; justify-content: space-between; padding: 1rem; border: 1px solid var(--border); border-radius: var(--radius-md); background: var(--surface);">
            <div class="applicant-info" style="display: flex; align-items: center; gap: 1rem; cursor: pointer" @click="navigateTo('profile:' + app.id)">
              <div class="chat-conv-avatar chat-conv-avatar--sm">{{ app.name.slice(0, 2).toUpperCase() }}</div>
              <strong>{{ app.name }}</strong>
            </div>
            <button class="button button--ghost" @click="navigateTo('chat:' + app.id)">
              <MessageSquare :size="16" />
              {{ $t('jobs.messageFreelancer') || 'Написать' }}
            </button>
          </div>
        </div>
      </div>
    </div>
  </section>
</template>
