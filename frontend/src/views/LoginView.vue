<script setup>
import { ref, inject } from 'vue';
import { LogIn, Mail, Lock, CircleAlert } from 'lucide-vue-next';

const navigateTo = inject('navigateTo');

const API_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8088/api';
const FALLBACK_API_URL = 'http://localhost:8080/api';

const email = ref('');
const password = ref('');
const error = ref('');
const loading = ref(false);

async function submit() {
  error.value = '';
  if (!email.value || !password.value) {
    error.value = 'Please fill in all fields';
    return;
  }
  loading.value = true;
  try {
    let res;
    const body = JSON.stringify({ email: email.value, password: password.value });
    const opts = { method: 'POST', headers: { 'Content-Type': 'application/json' }, body };
    try {
      res = await fetch(`${API_URL}/login`, opts);
    } catch {
      res = await fetch(`${FALLBACK_API_URL}/login`, opts);
    }
    const text = await res.text();
    let data = {};
    try { data = text ? JSON.parse(text) : {}; } catch { /* ignore */ }
    if (!res.ok) throw new Error(data.error || res.statusText || 'Login failed');
    const { token, id } = data;
    if (!token) throw new Error('Invalid server response');
    localStorage.setItem('fm_token', token);
    localStorage.setItem('fm_user_id', id);
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
          <LogIn :size="28" />
        </div>
        <h1>{{ $t('nav.login') }}</h1>
        <p>{{ $t('auth.loginSubtitle') }}</p>
      </div>

      <div v-if="error" class="auth-error">
        <CircleAlert :size="16" />
        {{ error }}
      </div>

      <form class="auth-form" @submit.prevent="submit">
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
              autocomplete="current-password"
              required
            />
          </div>
        </label>

        <button class="button button--primary auth-submit" type="submit" :disabled="loading">
          <LogIn :size="16" />
          {{ loading ? $t('common.loading') : $t('nav.login') }}
        </button>
      </form>

      <p class="auth-footer">
        {{ $t('auth.noAccount') }}
        <a href="#" @click.prevent="navigateTo('register')">
          {{ $t('nav.register') }}
        </a>
      </p>
    </div>
  </section>
</template>
