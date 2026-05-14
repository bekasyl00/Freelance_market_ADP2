<script setup>
import { computed, inject, onMounted, ref } from 'vue';
import { Plus, Save, Upload, LogOut, Camera, Trash2, Star } from 'lucide-vue-next';
import { marketplaceApi } from '../services/marketplace';

const props = defineProps({ userId: { type: String, default: '' } });

const appLogout = inject('logout');

const isOwnProfile = computed(() => !props.userId || props.userId === localStorage.getItem('fm_user_id'));

const profile = ref(null);
const skill = ref('');
const isSaving = ref(false);
const isSavingProfile = ref(false);
const avatarFile = ref(null);
const avatarPreview = ref(null);
const fileInput = ref(null);

// Review state (for client viewing freelancer)
const isViewerClient = ref(localStorage.getItem('fm_user_role') === 'client');
const canReview = computed(() => !isOwnProfile.value && isViewerClient.value && profile.value?.role === 'freelancer');
const reviewRating = ref(0);
const reviewHover = ref(0);
const reviewComment = ref('');
const isSubmittingReview = ref(false);
const reviewSent = ref(false);

async function submitProfileReview() {
  if (reviewRating.value < 1 || !props.userId) return;
  isSubmittingReview.value = true;
  try {
    await marketplaceApi.submitReview(props.userId, reviewRating.value, reviewComment.value);
    reviewSent.value = true;
    // Refresh profile to see updated rating
    profile.value = await marketplaceApi.getProfile(props.userId);
  } catch (err) {
    alert('Error submitting review');
  } finally {
    isSubmittingReview.value = false;
  }
}

function onFileChange(e) {
  const file = e.target.files[0];
  if (!file) return;
  avatarFile.value = file;
  avatarPreview.value = URL.createObjectURL(file);
}

const initials = computed(() => {
  if (!profile.value) return 'FM';
  return profile.value.name
    .split(' ')
    .map((part) => part[0])
    .join('')
    .slice(0, 2);
});

const avatarUrl = computed(() => {
  if (avatarPreview.value) return avatarPreview.value;
  if (profile.value?.avatar) return profile.value.avatar;
  return null;
});

onMounted(async () => {
  profile.value = await marketplaceApi.getProfile(props.userId);
  if (isOwnProfile.value && profile.value?.role) {
    localStorage.setItem('fm_user_role', profile.value.role);
  }
});

function addSkill() {
  const next = skill.value.trim();
  if (!next || profile.value.skills.includes(next)) return;
  profile.value.skills = [...profile.value.skills, next];
  skill.value = '';
}

function removeSkill(skillToRemove) {
  profile.value.skills = profile.value.skills.filter(s => s !== skillToRemove);
}

async function saveProfile() {
  if (!profile.value) return;
  isSavingProfile.value = true;
  try {
    profile.value = await marketplaceApi.updateProfile({ userId: profile.value.id, name: profile.value.name, avatar: profile.value.avatar });
  } catch (error) {
    console.warn(error);
  } finally {
    isSavingProfile.value = false;
  }
}

async function uploadAvatar() {
  if (!avatarFile.value) return;
  try {
    const res = await marketplaceApi.uploadAvatar(avatarFile.value);
    profile.value.avatar = res.url;
    avatarFile.value = null;
    avatarPreview.value = null;
  } catch (err) {
    console.warn(err);
  }
}

async function saveSkills() {
  if (!profile.value) return;
  isSaving.value = true;
  try {
    profile.value = await marketplaceApi.updateSkills(profile.value.id, profile.value.skills);
  } catch (error) {
    console.warn(error);
  } finally {
    isSaving.value = false;
  }
}

function logout() {
  if (appLogout) {
    appLogout();
  } else {
    localStorage.removeItem('fm_token');
    localStorage.removeItem('fm_user_id');
    window.location.reload();
  }
}
</script>

<template>
  <section v-if="profile" class="view-stack">
    <div class="section-heading">
      <div>
        <p class="eyebrow">{{ $t('profile.eyebrow') }}</p>
        <h1>{{ $t('profile.title') }}</h1>
      </div>
      <p>{{ $t('profile.subtitle') }}</p>
    </div>

    <div class="profile-layout">
      <section class="profile-summary">
        <div class="profile-avatar-section">
          <div class="avatar-wrapper" :style="{ cursor: isOwnProfile ? 'pointer' : 'default' }" @click="isOwnProfile && fileInput && fileInput.click()">
            <div v-if="avatarUrl" class="avatar avatar--image">
              <img :src="avatarUrl" alt="avatar" />
            </div>
            <div v-else class="avatar">{{ initials }}</div>
            <div v-if="isOwnProfile" class="avatar-overlay">
              <Camera :size="20" />
            </div>
          </div>
        </div>
        <input
          ref="fileInput"
          type="file"
          accept="image/*"
          class="sr-only"
          @change="onFileChange"
        />

        <div class="profile-info">
          <div class="profile-name-field">
            <h2 v-if="!isOwnProfile" class="profile-name-display" style="font-size: 1.5rem; font-weight: 700; margin: 0;">{{ profile.name }}</h2>
            <input v-else v-model="profile.name" class="profile-name-input" />
          </div>
          <p class="profile-role">
            <span class="profile-role-badge" :class="`profile-role-badge--${profile.role}`">
              {{ $t(`profile.${profile.role}`) }}
            </span>
          </p>
        </div>

        <div class="profile-actions" v-if="isOwnProfile">
          <button class="button button--ghost" @click="fileInput && fileInput.click()">
            <Upload :size="16" />
            {{ $t('profile.uploadAvatar') }}
          </button>
          <button
            v-if="avatarFile"
            class="button button--primary"
            @click="uploadAvatar"
          >
            <Save :size="16" />
            {{ $t('profile.saveAvatar') }}
          </button>
          <button
            class="button button--primary"
            :disabled="isSavingProfile"
            @click="saveProfile"
          >
            <Save :size="16" />
            {{ $t('common.save') }}
          </button>
          <button class="button button--danger" @click="appLogout">
            <LogOut :size="16" />
            {{ $t('nav.logout') || 'Logout' }}
          </button>
        </div>

        <dl>
          <div>
            <dt>{{ $t('profile.rating') }}</dt>
            <dd>
              <span class="profile-rating">
                ⭐ {{ profile.rating }}
              </span>
            </dd>
          </div>
          <div>
            <dt>{{ $t('profile.completed') }}</dt>
            <dd>{{ profile.completedJobs }}</dd>
          </div>
        </dl>

        <!-- Review section (client viewing freelancer) -->
        <div v-if="canReview" class="review-card">
          <h3 style="margin: 0 0 0.75rem 0; font-size: 1rem;">{{ $t('profile.leaveReview') || 'Leave a review' }}</h3>
          
          <div v-if="reviewSent" style="background: #ecfdf5; color: #065f46; border: 1px solid #6ee7b7; border-radius: 8px; padding: 0.5rem 1rem; font-size: 0.9rem;">
            ✅ {{ $t('profile.reviewSent') || 'Review submitted! Thank you.' }}
          </div>

          <template v-else>
            <div style="display: flex; gap: 4px; margin-bottom: 0.75rem;">
              <Star
                v-for="s in 5" :key="s" :size="28"
                :style="{ cursor: 'pointer', color: s <= (reviewHover || reviewRating) ? '#f59e0b' : '#d1d5db', fill: s <= (reviewHover || reviewRating) ? '#f59e0b' : 'none', transition: 'all 0.15s' }"
                @click="reviewRating = s"
                @mouseenter="reviewHover = s"
                @mouseleave="reviewHover = 0"
              />
              <span v-if="reviewRating" style="margin-left: 0.5rem; font-weight: 600; color: var(--text); align-self: center;">{{ reviewRating }}/5</span>
            </div>
            <textarea
              v-model="reviewComment"
              :placeholder="$t('profile.reviewPlaceholder') || 'Write your review here...'"
              rows="3"
              style="width: 100%; padding: 0.5rem 0.75rem; border: 1.5px solid var(--border, #e5e7eb); border-radius: 8px; font-family: inherit; font-size: 0.9rem; resize: vertical; background: var(--surface, #fff);"
            ></textarea>
            <button
              class="button button--primary"
              :disabled="isSubmittingReview || reviewRating < 1"
              @click="submitProfileReview"
              style="margin-top: 0.5rem;"
            >
              <Star :size="16" />
              {{ isSubmittingReview ? '...' : ($t('profile.submitReview') || 'Submit Review') }}
            </button>
          </template>
        </div>

      </section>

      <section class="skills-panel">
        <div class="section-title">
          <h2>{{ $t('profile.skills') }}</h2>
          <button v-if="isOwnProfile" class="button button--ghost" type="button" :disabled="isSaving" @click="saveSkills">
            <Save :size="16" />
            {{ $t('profile.updateSkills') }}
          </button>
        </div>

        <div class="skill-row skill-row--large" :class="{'skill-row--editable': isOwnProfile}">
          <span v-for="item in profile.skills" :key="item" class="skill-tag">
            {{ item }}
            <button v-if="isOwnProfile" class="skill-remove" type="button" @click="removeSkill(item)">
              <Trash2 :size="12" />
            </button>
          </span>
        </div>

        <form v-if="isOwnProfile" class="inline-form" @submit.prevent="addSkill">
          <input v-model="skill" :placeholder="$t('profile.addSkill')" type="text" />
          <button class="icon-button" type="submit" :aria-label="$t('profile.addSkill')">
            <Plus :size="18" />
          </button>
        </form>
      </section>
    </div>
  </section>
</template>

<style scoped>
.review-card {
  margin-top: 1.5rem;
  padding: 1.25rem;
  background: linear-gradient(135deg, var(--surface, #fff) 0%, var(--bg, #f9fafb) 100%);
  border: 1.5px solid var(--border, #e5e7eb);
  border-radius: 12px;
  box-shadow: 0 2px 12px rgba(0,0,0,0.04);
}
</style>
