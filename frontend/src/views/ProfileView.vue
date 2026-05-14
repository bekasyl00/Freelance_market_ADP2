<script setup>
import { computed, inject, onMounted, ref } from 'vue';
import { Plus, Save, Upload, LogOut, Camera, Trash2 } from 'lucide-vue-next';
import { marketplaceApi } from '../services/marketplace';

const appLogout = inject('logout');

const profile = ref(null);
const skill = ref('');
const isSaving = ref(false);
const isSavingProfile = ref(false);
const avatarFile = ref(null);
const avatarPreview = ref(null);
const fileInput = ref(null);

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
  profile.value = await marketplaceApi.getProfile();
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
          <div class="avatar-wrapper" @click="fileInput && fileInput.click()">
            <div v-if="avatarUrl" class="avatar avatar--image">
              <img :src="avatarUrl" alt="avatar" />
            </div>
            <div v-else class="avatar">{{ initials }}</div>
            <div class="avatar-overlay">
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
            <input v-model="profile.name" class="profile-name-input" />
          </div>
          <p class="profile-role">
            <span class="profile-role-badge" :class="`profile-role-badge--${profile.role}`">
              {{ $t(`profile.${profile.role}`) }}
            </span>
          </p>
        </div>

        <div class="profile-actions">
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
          <button class="button button--danger" @click="logout">
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
      </section>

      <section class="skills-panel">
        <div class="section-title">
          <h2>{{ $t('profile.skills') }}</h2>
          <button class="button button--ghost" type="button" :disabled="isSaving" @click="saveSkills">
            <Save :size="16" />
            {{ $t('profile.updateSkills') }}
          </button>
        </div>

        <div class="skill-row skill-row--large skill-row--editable">
          <span v-for="item in profile.skills" :key="item" class="skill-tag">
            {{ item }}
            <button class="skill-remove" type="button" @click="removeSkill(item)">
              <Trash2 :size="12" />
            </button>
          </span>
        </div>

        <form class="inline-form" @submit.prevent="addSkill">
          <input v-model="skill" :placeholder="$t('profile.addSkill')" type="text" />
          <button class="icon-button" type="submit" :aria-label="$t('profile.addSkill')">
            <Plus :size="18" />
          </button>
        </form>
      </section>
    </div>
  </section>
</template>
