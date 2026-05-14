<script setup>
import { ref, inject } from 'vue';
import { UserPlus, Mail, Lock, User, Briefcase, CircleAlert } from 'lucide-vue-next';

const navigateTo = inject('navigateTo');

const API_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8088/api';
const FALLBACK_API_URL = 'http://localhost:8080/api';

const email = ref('');
const password = ref('');
const name = ref('');
const role = ref('client');
const error = ref('');
const loading = ref(false);

async function submit() {
  error.value = '';
  if (!email.value || !password.value || !name.value) {
    error.value = 'Please fill in all fields';
    return;
  }
  loading.value = true;
  try {
    let res;
    const body = JSON.stringify({ email: email.value, password: password.value, name: name.value, role: role.value });
    const opts = { method: 'POST', headers: { 'Content-Type': 'application/json' }, body };
    try {
      res = await fetch(`${API_URL}/register`, opts);
    } catch {
      res = await fetch(`${FALLBACK_API_URL}/register`, opts);
    }
    const text = await res.text();
    let data = {};
    try { data = text ? JSON.parse(text) : {}; } catch { /* ignore */ }
    if (!res.ok) throw new Error(data.error || res.statusText || 'Register failed');
    const { token, id } = data;
    if (!token) throw new Error('Invalid server response');
    localStorage.setItem('fm_token', token);
    localStorage.setItem('fm_user_id', id);
    localStorage.setItem('fm_user_role', data.role || role.value);
    window.location.reload();
  } catch (e) {
    error.value = e.message;
  } finally {
    loading.value = false;
  }
}
</script>

<template>
  <section class="auth-page">
    <div class="auth-card">
      <div class="auth-card__header">
        <div class="auth-card__icon">
          <UserPlus :size="28" />
        </div>
        <h1>{{ $t('nav.register') }}</h1>
        <p>{{ $t('auth.registerSubtitle') }}</p>
      </div>

      <div v-if="error" class="auth-error">
        <CircleAlert :size="16" />
        {{ error }}
      </div>

      <form class="auth-form" @submit.prevent="submit">
        <label class="auth-field">
          <span class="auth-field__label">{{ $t('auth.name') }}</span>
          <div class="auth-field__input-wrap">
            <User :size="18" class="auth-field__icon" />
            <input
              v-model="name"
              type="text"
              :placeholder="$t('auth.namePlaceholder')"
              autocomplete="name"
              required
            />
          </div>
        </label>

        <label class="auth-field">
          <span class="auth-field__label">{{ $t('auth.email') }}</span>
          <div class="auth-field__input-wrap">
            <Mail :size="18" class="auth-field__icon" />
            <input
              v-model="email"
              type="email"
              :placeholder="$t('auth.emailPlaceholder')"
              autocomplete="email"
              required
            />
          </div>
        </label>

        <label class="auth-field">
          <span class="auth-field__label">{{ $t('auth.password') }}</span>
          <div class="auth-field__input-wrap">
            <Lock :size="18" class="auth-field__icon" />
            <input
              v-model="password"
              type="password"
              :placeholder="$t('auth.passwordPlaceholder')"
              autocomplete="new-password"
              required
            />
          </div>
        </label>

        <label class="auth-field">
          <span class="auth-field__label">{{ $t('auth.role') }}</span>
          <div class="auth-role-toggle">
            <button
              type="button"
              :class="['auth-role-option', { active: role === 'client' }]"
              @click="role = 'client'"
            >
              <Briefcase :size="16" />
              {{ $t('profile.client') }}
            </button>
            <button
              type="button"
              :class="['auth-role-option', { active: role === 'freelancer' }]"
              @click="role = 'freelancer'"
            >
              <User :size="16" />
              {{ $t('profile.freelancer') }}
            </button>
          </div>
        </label>

        <button class="button button--primary auth-submit" type="submit" :disabled="loading">
          <UserPlus :size="16" />
          {{ loading ? $t('common.loading') : $t('nav.register') }}
        </button>
      </form>

      <p class="auth-footer">
        {{ $t('auth.hasAccount') }}
        <a href="#" @click.prevent="navigateTo('login')">
          {{ $t('nav.login') }}
        </a>
      </p>
    </div>
  </section>
</template>
