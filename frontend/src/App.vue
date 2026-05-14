<script setup>
import { computed, provide, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { BriefcaseBusiness, CreditCard, LayoutDashboard, UserRound, MessageSquare, LogIn, LogOut, UserPlus } from 'lucide-vue-next';
import DashboardView from './views/DashboardView.vue';
import JobsView from './views/JobsView.vue';
import ProfileView from './views/ProfileView.vue';
import PaymentsView from './views/PaymentsView.vue';
import ChatView from './views/ChatView.vue';
import LoginView from './views/LoginView.vue';
import RegisterView from './views/RegisterView.vue';

const { locale, t } = useI18n();
const activeView = ref('dashboard');

const navItems = computed(() => [
  { key: 'dashboard', label: t('nav.dashboard'), icon: LayoutDashboard },
  { key: 'jobs', label: t('nav.jobs'), icon: BriefcaseBusiness },
  { key: 'profile', label: t('nav.profile'), icon: UserRound },
  { key: 'payments', label: t('nav.payments'), icon: CreditCard },
  { key: 'chat', label: t('nav.chat'), icon: MessageSquare },
]);

const viewMap = {
  dashboard: DashboardView,
  jobs: JobsView,
  profile: ProfileView,
  payments: PaymentsView,
  chat: ChatView,
  login: LoginView,
  register: RegisterView,
};

const token = ref(localStorage.getItem('fm_token'));
const isAuthenticated = computed(() => !!token.value);

function navigateTo(view) { activeView.value = view; }
function openLogin() { activeView.value = 'login'; }
function openRegister() { activeView.value = 'register'; }
function logout() {
  localStorage.removeItem('fm_token');
  localStorage.removeItem('fm_user_id');
  token.value = null;
  activeView.value = 'dashboard';
  window.location.reload();
}

// Provide navigation and logout to child components
provide('navigateTo', navigateTo);
provide('logout', logout);

// allow other tabs to update token state
window.addEventListener('storage', () => { token.value = localStorage.getItem('fm_token'); });

const activeComponent = computed(() => viewMap[activeView.value] || DashboardView);

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

      <component :is="activeComponent" />
    </main>
  </div>
</template>
