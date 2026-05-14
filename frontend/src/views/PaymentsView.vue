<script setup>
import { onMounted, ref } from 'vue';
import { CircleDollarSign, LockKeyhole, LockKeyholeOpen } from 'lucide-vue-next';
import MetricCard from '../components/MetricCard.vue';
import { marketplaceApi } from '../services/marketplace';

const payments = ref({
  available: 0,
  escrowed: 0,
  history: [],
});
const depositAmount = ref(500);
const isDepositing = ref(false);

onMounted(async () => {
  payments.value = await marketplaceApi.getPayments();
});

async function deposit() {
  if (depositAmount.value <= 0) return;
  isDepositing.value = true;
  try {
    payments.value = await marketplaceApi.deposit(Number(depositAmount.value));
  } catch (error) {
    console.warn(error);
  } finally {
    isDepositing.value = false;
  }
}
</script>

<template>
  <section class="view-stack">
    <div class="section-heading">
      <div>
        <p class="eyebrow">{{ $t('payments.eyebrow') }}</p>
        <h1>{{ $t('payments.title') }}</h1>
      </div>
      <p>{{ $t('payments.subtitle') }}</p>
    </div>

    <div class="metrics-grid metrics-grid--payments">
      <MetricCard :icon="CircleDollarSign" :label="$t('payments.available')" :value="`$${payments.available}`" :detail="$t('payments.deposit')" />
      <MetricCard :icon="LockKeyhole" :label="$t('payments.escrowed')" :value="`$${payments.escrowed}`" :detail="$t('payments.createEscrow')" />
      <MetricCard :icon="LockKeyholeOpen" :label="$t('payments.release')" value="$1,250" :detail="$t('common.completed')" />
    </div>

    <form class="job-form payment-form" @submit.prevent="deposit">
      <label>
        {{ $t('payments.deposit') }}
        <input v-model="depositAmount" min="1" step="50" type="number" />
      </label>
      <button class="button button--primary" type="submit" :disabled="isDepositing">
        <CircleDollarSign :size="16" />
        {{ $t('payments.deposit') }}
      </button>
    </form>

    <section class="table-panel">
      <div class="section-title">
        <h2>{{ $t('payments.history') }}</h2>
      </div>
      <table>
        <thead>
          <tr>
            <th>ID</th>
            <th>Type</th>
            <th>Amount</th>
            <th>{{ $t('common.status') }}</th>
            <th>Date</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="item in payments.history" :key="item.id">
            <td>{{ item.id }}</td>
            <td>{{ item.type }}</td>
            <td>${{ item.amount }}</td>
            <td>{{ item.status === 'completed' ? $t('common.completed') : $t('common.pending') }}</td>
            <td>{{ item.date }}</td>
          </tr>
        </tbody>
      </table>
    </section>
  </section>
</template>
