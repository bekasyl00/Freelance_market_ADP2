<script setup>
import { ref, inject, onMounted, computed } from 'vue';
import { ArrowLeft, CalendarDays, CircleDollarSign, Send, Users, User, CheckCircle, MessageSquare, Pencil, Trash2, ToggleLeft, ToggleRight, Save, X } from 'lucide-vue-next';
import { marketplaceApi } from '../services/marketplace';

const props = defineProps({ jobId: { type: String, default: '' } });
const navigateTo = inject('navigateTo');

const job = ref(null);
const applying = ref(false);
const applied = ref(false);
const error = ref('');
const isEditing = ref(false);
const editForm = ref({ title: '', description: '', budget: 0, deadline: '' });
const isSaving = ref(false);
const isDeleting = ref(false);
const isTogglingStatus = ref(false);

const userRole = computed(() => localStorage.getItem('fm_user_role') || 'freelancer');
const userId = computed(() => localStorage.getItem('fm_user_id') || '');
const isOwner = computed(() => job.value && job.value.clientId === userId.value);

onMounted(async () => {
  try {
    job.value = await marketplaceApi.getJobDetail(props.jobId);
  } catch {
    try {
      const all = await marketplaceApi.getJobs();
      job.value = all.find(j => j.id === props.jobId) || null;
    } catch {
      job.value = null;
    }
  }
  if (job.value && job.value.freelancers && userId.value) {
    if (job.value.freelancers.includes(userId.value)) {
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

function startEditing() {
  editForm.value = {
    title: job.value.title,
    description: job.value.description,
    budget: job.value.budget,
    deadline: job.value.deadline || '',
  };
  isEditing.value = true;
}

async function saveEdit() {
  isSaving.value = true;
  try {
    await marketplaceApi.updateJob(job.value.id, editForm.value);
    job.value.title = editForm.value.title;
    job.value.description = editForm.value.description;
    job.value.budget = editForm.value.budget;
    job.value.deadline = editForm.value.deadline;
    isEditing.value = false;
  } catch (e) {
    error.value = e.message;
  } finally {
    isSaving.value = false;
  }
}

async function toggleStatus() {
  if (!job.value) return;
  isTogglingStatus.value = true;
  const newStatus = job.value.status === 'open' ? 'in_progress' : 'open';
  try {
    await marketplaceApi.updateJob(job.value.id, { status: newStatus });
    job.value.status = newStatus;
  } catch (e) {
    error.value = e.message;
  } finally {
    isTogglingStatus.value = false;
  }
}

async function deleteJobAction() {
  if (!confirm('Вы уверены, что хотите удалить этот заказ?')) return;
  isDeleting.value = true;
  try {
    await marketplaceApi.deleteJob(job.value.id);
    navigateTo('jobs');
  } catch (e) {
    error.value = e.message;
  } finally {
    isDeleting.value = false;
  }
}

function statusLabel(status) {
  const map = { open: 'Open', in_progress: 'In Progress', inProgress: 'In Progress', completed: 'Completed' };
  return map[status] || status;
}

function statusClass(status) {
  if (status === 'in_progress') return 'inProgress';
  return status;
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
      <!-- Header with status -->
      <div class="job-detail__header">
        <div style="flex: 1; min-width: 0;">
          <p class="eyebrow">{{ job.client }}</p>
          <h1 v-if="!isEditing" class="job-detail__title">{{ job.title }}</h1>
          <input v-else v-model="editForm.title" class="input" style="font-size: 1.25rem; font-weight: 700; width: 100%;" />
        </div>
        <div style="display: flex; align-items: center; gap: 0.5rem;">
          <span class="status-pill" :class="`status-pill--${statusClass(job.status)}`">
            {{ statusLabel(job.status) }}
          </span>
          <!-- Status toggle for owner -->
          <button
            v-if="isOwner && job.status !== 'completed'"
            class="button button--ghost"
            :disabled="isTogglingStatus"
            @click="toggleStatus"
            :title="job.status === 'open' ? 'Set In Progress' : 'Set Open'"
            style="padding: 0.3rem 0.5rem;"
          >
            <component :is="job.status === 'open' ? ToggleLeft : ToggleRight" :size="20" />
          </button>
        </div>
      </div>

      <!-- Meta -->
      <div class="job-detail__meta">
        <div class="job-detail__meta-item">
          <CircleDollarSign :size="20" />
          <div>
            <span class="job-detail__meta-label">{{ $t('jobs.budget') }}</span>
            <strong v-if="!isEditing">${{ job.budget }}</strong>
            <input v-else v-model.number="editForm.budget" type="number" class="input" style="width: 100px;" />
          </div>
        </div>
        <div class="job-detail__meta-item">
          <CalendarDays :size="20" />
          <div>
            <span class="job-detail__meta-label">{{ $t('jobs.deadline') }}</span>
            <strong v-if="!isEditing">{{ job.deadline || '—' }}</strong>
            <input v-else v-model="editForm.deadline" type="date" class="input" />
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

      <!-- Description -->
      <div class="job-detail__body">
        <h2>{{ $t('jobs.description') || 'Description' }}</h2>
        <p v-if="!isEditing">{{ job.description }}</p>
        <textarea v-else v-model="editForm.description" class="input" rows="5" style="width: 100%; resize: vertical;"></textarea>
      </div>

      <!-- Skills -->
      <div class="job-detail__skills">
        <h3>{{ $t('jobs.skills') }}</h3>
        <div class="skill-row skill-row--large">
          <span v-for="skill in job.skills" :key="skill">{{ skill }}</span>
        </div>
      </div>

      <div v-if="error" class="auth-error" style="margin-top:16px">{{ error }}</div>

      <!-- Actions -->
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
          v-if="isOwner && (job.status === 'in_progress' || job.status === 'inProgress')"
          class="button button--primary"
          @click="completeJob"
        >
          <CheckCircle :size="16" />
          {{ $t('jobs.markComplete') || 'Mark as Complete' }}
        </button>

        <!-- Client edit/delete -->
        <template v-if="isOwner && !isEditing">
          <button class="button button--ghost" @click="startEditing">
            <Pencil :size="16" />
            {{ $t('common.edit') || 'Edit' }}
          </button>
          <button class="button button--danger" :disabled="isDeleting" @click="deleteJobAction">
            <Trash2 :size="16" />
            {{ isDeleting ? '...' : ($t('common.delete') || 'Delete') }}
          </button>
        </template>

        <!-- Editing save/cancel -->
        <template v-if="isEditing">
          <button class="button button--primary" :disabled="isSaving" @click="saveEdit">
            <Save :size="16" />
            {{ isSaving ? '...' : ($t('common.save') || 'Save') }}
          </button>
          <button class="button button--ghost" @click="isEditing = false">
            <X :size="16" />
            {{ $t('common.cancel') || 'Cancel' }}
          </button>
        </template>
      </div>

      <!-- Applicants (client only) -->
      <div v-if="isOwner && job.applicants && job.applicants.length > 0" class="job-detail__applicants" style="margin-top: 2rem; border-top: 1px solid var(--border); padding-top: 2rem;">
        <h3>{{ $t('jobs.applicants') || 'Applicants' }}</h3>
        <div class="applicants-list" style="display: flex; flex-direction: column; gap: 1rem; margin-top: 1rem;">
          <div v-for="app in job.applicants" :key="app.id" class="applicant-card" style="display: flex; align-items: center; justify-content: space-between; padding: 1rem; border: 1px solid var(--border); border-radius: var(--radius-md); background: var(--surface);">
            <div class="applicant-info" style="display: flex; align-items: center; gap: 1rem; cursor: pointer" @click="navigateTo('profile:' + app.id)">
              <div class="chat-conv-avatar chat-conv-avatar--sm">{{ app.name.slice(0, 2).toUpperCase() }}</div>
              <strong>{{ app.name }}</strong>
            </div>
            <button class="button button--ghost" @click="navigateTo('chat:' + app.id)">
              <MessageSquare :size="16" />
              {{ $t('jobs.messageFreelancer') || 'Message' }}
            </button>
          </div>
        </div>
      </div>
    </div>
  </section>
</template>
