<script setup>
import { onMounted, ref } from 'vue';
import { CircleDollarSign, LockKeyhole, LockKeyholeOpen, CreditCard, Wallet, ArrowUpRight, TrendingUp } from 'lucide-vue-next';
import MetricCard from '../components/MetricCard.vue';
import { marketplaceApi } from '../services/marketplace';

const isClient = ref(localStorage.getItem('fm_user_role') === 'client');
const isFreelancer = ref(localStorage.getItem('fm_user_role') === 'freelancer');

const payments = ref({
  available: 0,
  escrowed: 0,
  history: [],
});
const depositAmount = ref(500);
const cardNumber = ref('');
const cardExpiry = ref('');
const cardCvc = ref('');
const isDepositing = ref(false);
const depositSuccess = ref(false);

onMounted(async () => {
  payments.value = await marketplaceApi.getPayments();
});

function formatCardNumber(e) {
  let v = e.target.value.replace(/\D/g, '').slice(0, 16);
  cardNumber.value = v.replace(/(\d{4})(?=\d)/g, '$1 ');
}

function formatExpiry(e) {
  let v = e.target.value.replace(/\D/g, '').slice(0, 4);
  if (v.length > 2) v = v.slice(0, 2) + '/' + v.slice(2);
  cardExpiry.value = v;
}

async function deposit() {
  if (depositAmount.value <= 0) return;
  isDepositing.value = true;
  depositSuccess.value = false;
  try {
    payments.value = await marketplaceApi.deposit(Number(depositAmount.value));
    depositSuccess.value = true;
    setTimeout(() => { depositSuccess.value = false; }, 3000);
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
        <p class="eyebrow">{{ isFreelancer ? ($t('payments.walletEyebrow') || 'Earnings') : $t('payments.eyebrow') }}</p>
        <h1>{{ isFreelancer ? ($t('payments.walletTitle') || 'Freelancer Wallet') : $t('payments.title') }}</h1>
      </div>
      <p>{{ isFreelancer ? ($t('payments.walletSubtitle') || 'Your earnings and transaction history.') : $t('payments.subtitle') }}</p>
    </div>

    <div class="metrics-grid metrics-grid--payments">
      <MetricCard :icon="CircleDollarSign" :label="$t('payments.available')" :value="`$${payments.available}`" :detail="isFreelancer ? ($t('payments.earned') || 'Earned from clients') : $t('payments.deposit')" />
      <MetricCard :icon="LockKeyhole" :label="$t('payments.escrowed')" :value="`$${payments.escrowed}`" :detail="$t('payments.createEscrow')" />
      <MetricCard v-if="isFreelancer" :icon="TrendingUp" :label="$t('payments.totalEarned') || 'Total earned'" :value="`$${payments.available + payments.escrowed}`" :detail="$t('payments.earnings') || 'Your earnings'" />
      <MetricCard v-else :icon="LockKeyholeOpen" :label="$t('payments.release')" value="—" :detail="$t('common.completed')" />
    </div>

    <!-- Freelancer earnings info -->
    <div v-if="isFreelancer" class="payment-card" style="max-width: 460px;">
      <div class="payment-card__header">
        <TrendingUp :size="22" />
        <h3>{{ $t('payments.myEarnings') || 'My Earnings' }}</h3>
      </div>
      <p style="color: var(--text-secondary, #666); font-size: 0.95rem; line-height: 1.6; margin: 0;">
        {{ $t('payments.freelancerInfo') || 'Your balance is updated automatically when clients send you payments. All received transfers appear in the transaction history below.' }}
      </p>
    </div>

    <!-- Deposit card (clients only) -->
    <div v-if="isClient" class="payment-card">
      <div class="payment-card__header">
        <Wallet :size="22" />
        <h3>{{ $t('payments.topUpWallet') || 'Top up wallet' }}</h3>
      </div>

      <div v-if="depositSuccess" class="payment-card__success">
        ✅ {{ $t('payments.depositSuccess') || 'Deposit successful!' }}
      </div>

      <form @submit.prevent="deposit" class="payment-card__form">
        <div class="payment-card__row">
          <label class="payment-card__field">
            <span>{{ $t('payments.amount') || 'Amount' }} ($)</span>
            <div class="payment-card__input-group">
              <CircleDollarSign :size="16" class="payment-card__icon" />
              <input v-model="depositAmount" min="1" step="50" type="number" required />
            </div>
          </label>
        </div>

        <div class="payment-card__row">
          <label class="payment-card__field payment-card__field--full">
            <span>{{ $t('payments.cardNumber') || 'Card number' }}</span>
            <div class="payment-card__input-group">
              <CreditCard :size="16" class="payment-card__icon" />
              <input :value="cardNumber" @input="formatCardNumber" type="text" placeholder="0000 0000 0000 0000" maxlength="19" required />
            </div>
          </label>
        </div>

        <div class="payment-card__row payment-card__row--split">
          <label class="payment-card__field">
            <span>{{ $t('payments.expiry') || 'Expiry' }}</span>
            <div class="payment-card__input-group">
              <input :value="cardExpiry" @input="formatExpiry" type="text" placeholder="MM/YY" maxlength="5" required />
            </div>
          </label>
          <label class="payment-card__field">
            <span>CVC</span>
            <div class="payment-card__input-group">
              <input v-model="cardCvc" type="password" placeholder="•••" maxlength="3" required />
            </div>
          </label>
        </div>

        <button class="button button--primary payment-card__submit" type="submit" :disabled="isDepositing">
          <ArrowUpRight :size="16" />
          {{ isDepositing ? ($t('common.loading') || 'Loading...') : ($t('payments.deposit') || 'Deposit') }}
        </button>
      </form>
    </div>

    <!-- Transaction history -->
    <section class="table-panel">
      <div class="section-title">
        <h2>{{ $t('payments.history') }}</h2>
      </div>
      <table>
        <thead>
          <tr>
            <th>{{ $t('payments.type') || 'Type' }}</th>
            <th>{{ $t('payments.amount') || 'Amount' }}</th>
            <th>{{ $t('common.status') }}</th>
            <th>{{ $t('payments.date') || 'Date' }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="item in payments.history" :key="item.id">
            <td>
              <span class="payment-type-badge" :class="'payment-type-badge--' + item.type">
                {{ item.type === 'deposit' ? '💰 Deposit' : item.type === 'transfer_out' ? '📤 Sent' : item.type === 'transfer_in' ? '📥 Received' : item.type }}
              </span>
            </td>
            <td class="payment-amount">${{ item.amount }}</td>
            <td>
              <span class="status-dot" :class="item.status === 'completed' ? 'status-dot--completed' : 'status-dot--pending'"></span>
              {{ item.status === 'completed' ? $t('common.completed') : $t('common.pending') }}
            </td>
            <td>{{ item.date }}</td>
          </tr>
        </tbody>
      </table>
    </section>
  </section>
</template>

<style scoped>
.payment-card {
  max-width: 460px;
  background: linear-gradient(135deg, var(--surface) 0%, var(--bg) 100%);
  border: 1px solid var(--border);
  border-radius: var(--radius-lg, 12px);
  padding: 1.75rem;
  box-shadow: 0 4px 24px rgba(0,0,0,0.06);
}

.payment-card__header {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  margin-bottom: 1.25rem;
  color: var(--primary, #0d9488);
}
.payment-card__header h3 {
  margin: 0;
  font-size: 1.15rem;
  font-weight: 700;
}

.payment-card__success {
  background: #ecfdf5;
  color: #065f46;
  border: 1px solid #6ee7b7;
  border-radius: var(--radius-md, 8px);
  padding: 0.5rem 1rem;
  margin-bottom: 1rem;
  font-size: 0.9rem;
  font-weight: 500;
}

.payment-card__form {
  display: flex;
  flex-direction: column;
  gap: 0.9rem;
}

.payment-card__row {
  display: flex;
  gap: 0.75rem;
}
.payment-card__row--split {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 0.75rem;
}

.payment-card__field {
  display: flex;
  flex-direction: column;
  gap: 0.3rem;
  flex: 1;
}
.payment-card__field--full {
  width: 100%;
}
.payment-card__field span {
  font-size: 0.8rem;
  font-weight: 600;
  color: var(--text-secondary, #666);
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.payment-card__input-group {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  background: var(--surface, #fff);
  border: 1.5px solid var(--border, #e5e7eb);
  border-radius: var(--radius-md, 8px);
  padding: 0.55rem 0.75rem;
  transition: border-color 0.2s;
}
.payment-card__input-group:focus-within {
  border-color: var(--primary, #0d9488);
  box-shadow: 0 0 0 3px rgba(13,148,136,0.1);
}
.payment-card__input-group input {
  border: none;
  outline: none;
  background: transparent;
  flex: 1;
  font-size: 0.95rem;
  color: var(--text, #1f2937);
  font-family: inherit;
}
.payment-card__icon {
  color: var(--text-secondary, #9ca3af);
  flex-shrink: 0;
}

.payment-card__submit {
  margin-top: 0.5rem;
  padding: 0.65rem 1.25rem;
  font-size: 0.95rem;
  font-weight: 600;
  display: flex;
  align-items: center;
  gap: 0.5rem;
  justify-content: center;
}

/* Type badges */
.payment-type-badge {
  font-size: 0.85rem;
  font-weight: 500;
  padding: 0.2rem 0.5rem;
  border-radius: 6px;
  display: inline-block;
}
.payment-type-badge--deposit { background: #ecfdf5; color: #065f46; }
.payment-type-badge--transfer_out { background: #fef3c7; color: #92400e; }
.payment-type-badge--transfer_in { background: #dbeafe; color: #1e40af; }

.payment-amount {
  font-weight: 600;
  font-variant-numeric: tabular-nums;
}

/* Status dot */
.status-dot {
  display: inline-block;
  width: 8px;
  height: 8px;
  border-radius: 50%;
  margin-right: 0.3rem;
}
.status-dot--completed { background: #10b981; }
.status-dot--pending { background: #f59e0b; }
</style>
