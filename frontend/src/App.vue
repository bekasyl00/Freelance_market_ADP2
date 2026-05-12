<script setup>
import { computed, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { BriefcaseBusiness, CreditCard, LayoutDashboard, UserRound } from 'lucide-vue-next';
import DashboardView from './views/DashboardView.vue';
import JobsView from './views/JobsView.vue';
import ProfileView from './views/ProfileView.vue';
import PaymentsView from './views/PaymentsView.vue';

const { locale, t } = useI18n();
const activeView = ref('dashboard');

const navItems = computed(() => [
  { key: 'dashboard', label: t('nav.dashboard'), icon: LayoutDashboard, component: DashboardView },
  { key: 'jobs', label: t('nav.jobs'), icon: BriefcaseBusiness, component: JobsView },
  { key: 'profile', label: t('nav.profile'), icon: UserRound, component: ProfileView },
  { key: 'payments', label: t('nav.payments'), icon: CreditCard, component: PaymentsView },
]);

const activeComponent = computed(() => navItems.value.find((item) => item.key === activeView.value)?.component || DashboardView);

watch(locale, (value) => {
  localStorage.setItem('freelance-market-locale', value);
});
</script>

<template>
  <div class="app-shell">
    <aside class="sidebar">
      <div class="brand-block">
        <div class="brand-mark">FM</div>
        <div>
          <strong>{{ $t('app.name') }}</strong>
          <span>{{ $t('app.tagline') }}</span>
        </div>
      </div>

      <nav class="nav-list" aria-label="Main navigation">
        <button
          v-for="item in navItems"
          :key="item.key"
          :class="{ active: activeView === item.key }"
          type="button"
          @click="activeView = item.key"
        >
          <component :is="item.icon" :size="18" />
          {{ item.label }}
        </button>
      </nav>
    </aside>

    <main class="main-area">
      <header class="topbar">
        <div>
          <strong>{{ $t('app.name') }}</strong>
          <span>{{ $t('app.topline') }}</span>
        </div>
        <label class="locale-switch">
          {{ $t('app.language') }}
          <select v-model="locale">
            <option value="en">English</option>
            <option value="ru">Русский</option>
            <option value="kk">Қазақша</option>
          </select>
        </label>
      </header>

      <component :is="activeComponent" />
    </main>
  </div>
</template>
