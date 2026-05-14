<script setup>
import { ref, inject } from 'vue';
import { Plus, ArrowLeft, CalendarDays, CircleDollarSign, FileText, Tag, Type } from 'lucide-vue-next';
import { marketplaceApi } from '../services/marketplace';

const navigateTo = inject('navigateTo');
const isPublishing = ref(false);
const error = ref('');
const newJob = ref({
  title: '',
  budget: 750,
  description: '',
  skills: '',
  deadline: '',
});

async function publishJob() {
  error.value = '';
  if (!newJob.value.title.trim()) { error.value = 'Title is required'; return; }
  if (!newJob.value.budget || newJob.value.budget <= 0) { error.value = 'Budget must be positive'; return; }

  isPublishing.value = true;
  const payload = {
    clientId: localStorage.getItem('fm_user_id') || '',
    title: newJob.value.title,
    budget: Number(newJob.value.budget) || 0,
    description: newJob.value.description || 'Client is ready to discuss scope and deliverables.',
    skills: newJob.value.skills ? newJob.value.skills.split(',').map(s => s.trim()).filter(Boolean) : [],
    deadline: newJob.value.deadline || '',
  };

  try {
    await marketplaceApi.createJob(payload);
    navigateTo('jobs');
  } catch (err) {
    error.value = err.message || 'Failed to create job';
  } finally {
    isPublishing.value = false;
  }
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
        <p class="eyebrow">{{ $t('jobs.createEyebrow') || 'New project' }}</p>
        <h1>{{ $t('jobs.create') }}</h1>
      </div>
      <p>{{ $t('jobs.createSubtitle') || 'Fill in the details to publish a new job for freelancers.' }}</p>
    </div>

    <div v-if="error" class="auth-error">
      {{ error }}
    </div>

    <form class="create-job-form" @submit.prevent="publishJob">
      <label class="auth-field">
        <span class="auth-field__label">{{ $t('jobs.newTitle') }}</span>
        <div class="auth-field__input-wrap">
          <Type :size="18" class="auth-field__icon" />
          <input v-model="newJob.title" type="text" :placeholder="$t('jobs.titlePlaceholder') || 'e.g. Build a landing page'" required />
        </div>
      </label>

      <div class="create-job-row">
        <label class="auth-field">
          <span class="auth-field__label">{{ $t('jobs.newBudget') }}</span>
          <div class="auth-field__input-wrap">
            <CircleDollarSign :size="18" class="auth-field__icon" />
            <input v-model="newJob.budget" type="number" min="50" step="50" required />
          </div>
        </label>

        <label class="auth-field">
          <span class="auth-field__label">{{ $t('jobs.deadline') }}</span>
          <div class="auth-field__input-wrap">
            <CalendarDays :size="18" class="auth-field__icon" />
            <input v-model="newJob.deadline" type="date" />
          </div>
        </label>
      </div>

      <label class="auth-field">
        <span class="auth-field__label">{{ $t('jobs.skills') }}</span>
        <div class="auth-field__input-wrap">
          <Tag :size="18" class="auth-field__icon" />
          <input v-model="newJob.skills" type="text" :placeholder="$t('jobs.skillsPlaceholder') || 'e.g. Web Design, Frontend, SEO'" />
        </div>
      </label>

      <label class="auth-field">
        <span class="auth-field__label">{{ $t('jobs.details') }}</span>
        <div class="create-job-textarea-wrap">
          <FileText :size="18" class="auth-field__icon create-job-textarea-icon" />
          <textarea v-model="newJob.description" rows="5" :placeholder="$t('jobs.descPlaceholder') || 'Describe the project scope, deliverables, and expectations...'"></textarea>
        </div>
      </label>

      <button class="button button--primary auth-submit" type="submit" :disabled="isPublishing">
        <Plus :size="16" />
        {{ isPublishing ? ($t('common.loading') || 'Loading...') : $t('jobs.publish') }}
      </button>
    </form>
  </section>
</template>
