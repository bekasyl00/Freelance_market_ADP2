<script setup>
import { onMounted, ref } from 'vue';
import { BriefcaseBusiness, CircleDollarSign, Send, ShieldCheck, Star, UsersRound } from 'lucide-vue-next';
import MetricCard from '../components/MetricCard.vue';
import { marketplaceApi } from '../services/marketplace';

const summary = ref({
  activeJobs: 0,
  escrowBalance: 0,
  proposals: 0,
  rating: 0,
});

onMounted(async () => {
  summary.value = await marketplaceApi.getSummary();
});
</script>

<template>
  <section class="view-stack">
    <div class="section-heading">
      <div>
        <p class="eyebrow">{{ $t('dashboard.eyebrow') }}</p>
        <h1>{{ $t('dashboard.title') }}</h1>
      </div>
      <p>{{ $t('dashboard.subtitle') }}</p>
    </div>

    <div class="metrics-grid">
      <MetricCard :icon="BriefcaseBusiness" :label="$t('dashboard.activeJobs')" :value="summary.activeJobs" detail="+12%" />
      <MetricCard :icon="CircleDollarSign" :label="$t('dashboard.escrow')" :value="`$${summary.escrowBalance.toLocaleString()}`" detail="escrow" />
      <MetricCard :icon="Send" :label="$t('dashboard.proposals')" :value="summary.proposals" :detail="$t('common.pending')" />
      <MetricCard :icon="Star" :label="$t('dashboard.rating')" :value="summary.rating" :detail="$t('profile.freelancer')" />
    </div>

    <div class="split-layout">
      <section class="activity-panel">
        <div class="section-title">
          <h2>{{ $t('dashboard.popularCategories') }}</h2>
          <span>{{ $t('dashboard.updatedToday') }}</span>
        </div>
        <div class="category-grid">
          <article class="category-tile">
            <BriefcaseBusiness :size="20" />
            <h3>{{ $t('dashboard.categoryDevelopment') }}</h3>
            <p>128 {{ $t('dashboard.openProjects') }}</p>
          </article>
          <article class="category-tile">
            <UsersRound :size="20" />
            <h3>{{ $t('dashboard.categoryDesign') }}</h3>
            <p>74 {{ $t('dashboard.openProjects') }}</p>
          </article>
          <article class="category-tile">
            <ShieldCheck :size="20" />
            <h3>{{ $t('dashboard.categoryConsulting') }}</h3>
            <p>42 {{ $t('dashboard.openProjects') }}</p>
          </article>
        </div>
      </section>

      <aside class="note-panel">
        <h2>{{ $t('dashboard.howItWorks') }}</h2>
        <ol class="steps-list">
          <li>{{ $t('dashboard.stepPost') }}</li>
          <li>{{ $t('dashboard.stepApply') }}</li>
          <li>{{ $t('dashboard.stepPay') }}</li>
        </ol>
      </aside>
    </div>
  </section>
</template>
