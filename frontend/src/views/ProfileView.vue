<script setup>
import { computed, onMounted, ref } from 'vue';
import { Plus, Save } from 'lucide-vue-next';
import { marketplaceApi } from '../services/marketplace';

const profile = ref(null);
const skill = ref('');

const initials = computed(() => {
  if (!profile.value) return 'FM';
  return profile.value.name
    .split(' ')
    .map((part) => part[0])
    .join('')
    .slice(0, 2);
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
        <div class="avatar">{{ initials }}</div>
        <div>
          <h2>{{ profile.name }}</h2>
          <p>{{ $t('profile.role') }}: {{ $t(`profile.${profile.role}`) }}</p>
        </div>
        <dl>
          <div>
            <dt>{{ $t('profile.rating') }}</dt>
            <dd>{{ profile.rating }}</dd>
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
          <button class="button button--ghost" type="button">
            <Save :size="16" />
            {{ $t('profile.updateSkills') }}
          </button>
        </div>

        <div class="skill-row skill-row--large">
          <span v-for="item in profile.skills" :key="item">{{ item }}</span>
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
