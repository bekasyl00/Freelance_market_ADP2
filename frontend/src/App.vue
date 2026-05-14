<script setup>
import { computed, provide, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { BriefcaseBusiness, CreditCard, LayoutDashboard, UserRound, MessageSquare, LogIn, LogOut, UserPlus } from 'lucide-vue-next';
import DashboardView from './views/DashboardView.vue';
import JobsView from './views/JobsView.vue';
import CreateJobView from './views/CreateJobView.vue';
import JobDetailView from './views/JobDetailView.vue';
import ProfileView from './views/ProfileView.vue';
import PaymentsView from './views/PaymentsView.vue';
import ChatView from './views/ChatView.vue';
import LoginView from './views/LoginView.vue';
import RegisterView from './views/RegisterView.vue';

const { locale, t } = useI18n();
const activeView = ref('dashboard');

const token = ref(localStorage.getItem('fm_token'));
const userRole = ref(localStorage.getItem('fm_user_role') || '');
const isAuthenticated = computed(() => !!token.value);
const isClient = computed(() => userRole.value === 'client');
const isFreelancer = computed(() => userRole.value === 'freelancer');

// Nav items filtered by role
const navItems = computed(() => {
  const items = [
    { key: 'dashboard', label: t('nav.dashboard'), icon: LayoutDashboard },
    { key: 'jobs', label: t('nav.jobs'), icon: BriefcaseBusiness },
  ];
  if (isAuthenticated.value) {
    items.push({ key: 'profile', label: t('nav.profile'), icon: UserRound });
    if (!isFreelancer.value) {
      items.push({ key: 'payments', label: t('nav.payments'), icon: CreditCard });
    }
    items.push({ key: 'chat', label: t('nav.chat'), icon: MessageSquare });
  }
  return items;
});

const viewMap = {
  dashboard: DashboardView,
  jobs: JobsView,
  createJob: CreateJobView,
  profile: ProfileView,
  payments: PaymentsView,
  chat: ChatView,
  login: LoginView,
  register: RegisterView,
};

// Support parameterized views like 'jobDetail:job-101' and 'chat:user-101'
const activeComponent = computed(() => {
  const view = activeView.value;
  if (view.startsWith('jobDetail:')) return JobDetailView;
  if (view.startsWith('chat:')) return ChatView;
  if (view.startsWith('profile:')) return ProfileView;
  return viewMap[view] || DashboardView;
});

const activeProps = computed(() => {
  const view = activeView.value;
  if (view.startsWith('jobDetail:')) {
    return { jobId: view.replace('jobDetail:', '') };
  }
  if (view.startsWith('chat:')) {
    const raw = view.replace('chat:', '');
    const parts = raw.split('|');
    return { partnerId: parts[0], autoMsg: parts[1] || '' };
  }
  if (view.startsWith('profile:')) {
    return { userId: view.replace('profile:', '') };
  }
  return {};
});

// Highlight the nav item for sub-views
const activeNavKey = computed(() => {
  const view = activeView.value;
  if (view === 'createJob' || view.startsWith('jobDetail:')) return 'jobs';
  if (view.startsWith('chat:')) return 'chat';
  return view;
});

function navigateTo(view) { activeView.value = view; }
function openLogin() { activeView.value = 'login'; }
function openRegister() { activeView.value = 'register'; }
function logout() {
  localStorage.removeItem('fm_token');
  localStorage.removeItem('fm_user_id');
  localStorage.removeItem('fm_user_role');
  token.value = null;
  userRole.value = '';
  activeView.value = 'dashboard';
  window.location.reload();
}

provide('navigateTo', navigateTo);
provide('logout', logout);

window.addEventListener('storage', () => {
  token.value = localStorage.getItem('fm_token');
  userRole.value = localStorage.getItem('fm_user_role') || '';
});

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
          :class="{ active: activeNavKey === item.key }"
          type="button"
          @click="activeView = item.key"
        >
          <component :is="item.icon" :size="18" />
          {{ item.label }}
        </button>
      </nav>

      <div class="sidebar-auth">
        <template v-if="!isAuthenticated">
          <button class="sidebar-auth-btn sidebar-auth-btn--login" @click="openLogin">
            <LogIn :size="16" />
            {{ $t('nav.login') }}
          </button>
          <button class="sidebar-auth-btn sidebar-auth-btn--register" @click="openRegister">
            <UserPlus :size="16" />
            {{ $t('nav.register') }}
          </button>
        </template>
        <button v-else class="sidebar-auth-btn sidebar-auth-btn--logout" @click="logout">
          <LogOut :size="16" />
          {{ $t('nav.logout') || 'Logout' }}
        </button>
      </div>
    </aside>

    <main class="main-area">
      <header class="topbar">
        <div>
          <strong>{{ $t('app.name') }}</strong>
          <span>{{ $t('app.topline') }}</span>
        </div>
        <div class="topbar-actions">
          <label class="locale-switch">
            {{ $t('app.language') }}
            <select v-model="locale">
              <option value="en">English</option>
              <option value="ru">Русский</option>
              <option value="kk">Қазақша</option>
            </select>
          </label>
          <div class="topbar-auth">
            <button v-if="!isAuthenticated" class="button button--ghost" @click="openLogin">
              <LogIn :size="16" />
              {{ $t('nav.login') }}
            </button>
            <button v-if="!isAuthenticated" class="button button--primary" @click="openRegister">
              <UserPlus :size="16" />
              {{ $t('nav.register') }}
            </button>
            <button v-if="isAuthenticated" class="button button--danger" @click="logout">
              <LogOut :size="16" />
              {{ $t('nav.logout') || 'Logout' }}
            </button>
          </div>
        </div>
      </header>

      <component :is="activeComponent" v-bind="activeProps" />
    </main>
  </div>
</template>
