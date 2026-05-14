<script setup>
import { ref, onMounted, onBeforeUnmount, nextTick, computed, watch } from 'vue';
import { Send, MessageSquare, Users, Search, AlertCircle, Wifi, WifiOff, CircleDollarSign, Star } from 'lucide-vue-next';
import { marketplaceApi } from '../services/marketplace';

const props = defineProps({ partnerId: { type: String, default: '' }, autoMsg: { type: String, default: '' } });

const ws = ref(null);
const messages = ref([]);
const to = ref(props.partnerId);
const content = ref('');
const error = ref('');
const connected = ref(false);
const searchQuery = ref('');
const messagesContainer = ref(null);
const currentUserId = ref(localStorage.getItem('fm_user_id') || '');
const isClient = ref(localStorage.getItem('fm_user_role') === 'client');
const showPaymentModal = ref(false);
const paymentAmount = ref(50);
const isPaying = ref(false);
const showReviewModal = ref(false);
const reviewRating = ref(0);
const reviewHover = ref(0);
const reviewComment = ref('');
const isSubmittingReview = ref(false);

async function payFreelancer() {
  if (paymentAmount.value <= 0) return;
  isPaying.value = true;
  try {
    await marketplaceApi.transfer(to.value, Number(paymentAmount.value));
    showPaymentModal.value = false;
    alert('Оплата успешно отправлена!');
  } catch (err) {
    alert('Ошибка при оплате. Проверьте баланс.');
  } finally {
    isPaying.value = false;
  }
}

async function submitReview() {
  if (reviewRating.value < 1) return;
  isSubmittingReview.value = true;
  try {
    await marketplaceApi.submitReview(to.value, reviewRating.value, reviewComment.value);
    showReviewModal.value = false;
    reviewRating.value = 0;
    reviewComment.value = '';
    alert('Отзыв отправлен!');
  } catch (err) {
    alert('Ошибка при отправке отзыва.');
  } finally {
    isSubmittingReview.value = false;
  }
}

function navigateTo(view) {
  window.location.hash = view;
}

// Group messages by conversation partner
const conversations = computed(() => {
  const convMap = {};
  messages.value.forEach(m => {
    const partner = m.from === 'me' ? (m.to || to.value) : m.from;
    if (!partner) return;
    if (!convMap[partner]) {
      convMap[partner] = { id: partner, name: m.partnerName || partner, messages: [], lastMessage: '', lastTime: new Date() };
    }
    if (m.partnerName) convMap[partner].name = m.partnerName;
    convMap[partner].messages.push(m);
    convMap[partner].lastMessage = m.content;
  });
  return Object.values(convMap);
});

const activeConversation = computed(() => {
  if (!to.value) return [];
  return messages.value.filter(m =>
    !m.isSystem &&
    ((m.from === 'me' && (m.to === to.value || !m.to)) || m.from === to.value)
  );
});

function scrollToBottom() {
  nextTick(() => {
    if (messagesContainer.value) {
      messagesContainer.value.scrollTop = messagesContainer.value.scrollHeight;
    }
  });
}

function addMessage(m) {
  messages.value.push(m);
  scrollToBottom();
}

onMounted(async () => {
  const token = localStorage.getItem('fm_token');
  if (!token) { error.value = 'not_authenticated'; return; }

  // Fetch message history
  messages.value = await marketplaceApi.getMessages();
  
  if (to.value && !messages.value.some(m => m.from === to.value || m.to === to.value)) {
    const profile = await marketplaceApi.getProfile(to.value);
    if (profile && profile.name) {
      // Add a dummy system message just to register the name in the conversation list, or we can just push an empty message
      messages.value.push({
        from: to.value,
        content: '',
        partnerName: profile.name,
        time: new Date(),
        isSystem: true
      });
    }
  }

  scrollToBottom();

  const baseUrl = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8088/api';
  const apiHost = baseUrl.replace(/^https?:\/\//, '').replace(/\/api$/, '');
  const protocol = baseUrl.startsWith('https') ? 'wss' : 'ws';

  const wsUrl = `${protocol}://${apiHost}/ws/chat?token=${token}`;

  try {
    ws.value = new WebSocket(wsUrl);
    ws.value.addEventListener('open', () => { connected.value = true; });
    ws.value.addEventListener('close', () => { connected.value = false; });
    ws.value.addEventListener('error', () => { connected.value = false; });
    ws.value.addEventListener('message', (e) => {
      try {
        const data = JSON.parse(e.data);
        addMessage({ from: data.from, content: data.content, partnerName: data.partnerName, time: new Date() });
      } catch (err) { /* ignore */ }
    });
  } catch {
    error.value = 'connection_failed';
  }
});

let autoMsgSent = false;
watch(connected, (isConn) => {
  if (isConn && props.autoMsg && !autoMsgSent) {
    autoMsgSent = true;
    setTimeout(() => {
      const link = `${window.location.origin}/#jobDetail:${props.autoMsg}`;
      content.value = `Здравствуйте! Я откликнулся на ваш заказ. Ссылка на заказ: ${link}\nБуду рад обсудить детали!`;
      send();
    }, 500); // small delay to ensure rendering
  }
});

onBeforeUnmount(() => { if (ws.value) ws.value.close(); });

function send() {
  if (!to.value.trim() || !content.value.trim()) return;
  if (!ws.value || ws.value.readyState !== WebSocket.OPEN) {
    error.value = 'not_connected';
    return;
  }
  const payload = { to: to.value, content: content.value };
  ws.value.send(JSON.stringify(payload));
  addMessage({ from: 'me', to: to.value, content: content.value, time: new Date() });
  content.value = '';
}

function selectConversation(partnerId) {
  to.value = partnerId;
}

function formatTime(date) {
  if (!date) return '';
  const d = new Date(date);
  return d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
}
</script>

<template>
  <section class="view-stack">
    <div class="section-heading">
      <div>
        <p class="eyebrow">{{ $t('chat.eyebrow') || 'Communication' }}</p>
        <h1>{{ $t('nav.chat') }}</h1>
      </div>
      <p>{{ $t('chat.subtitle') || 'Send messages to other users in real time.' }}</p>
    </div>

    <!-- Not authenticated state -->
    <div v-if="error === 'not_authenticated'" class="chat-empty-state">
      <div class="chat-empty-state__icon">
        <AlertCircle :size="48" />
      </div>
      <h3>{{ $t('chat.loginRequired') || 'Login required' }}</h3>
      <p>{{ $t('chat.loginMessage') || 'Please log in to use the chat.' }}</p>
    </div>

    <!-- Chat interface -->
    <div v-else class="chat-container">
      <!-- Sidebar with conversations -->
      <aside class="chat-sidebar">
        <div class="chat-sidebar__header">
          <h3>
            <Users :size="18" />
            {{ $t('chat.conversations') || 'Conversations' }}
          </h3>
          <span class="chat-status" :class="connected ? 'chat-status--online' : 'chat-status--offline'">
            <component :is="connected ? Wifi : WifiOff" :size="12" />
            {{ connected ? ($t('common.online') || 'Online') : 'Offline' }}
          </span>
        </div>



        <!-- Conversation list -->
        <div class="chat-conv-list">
          <button
            v-for="conv in conversations"
            :key="conv.id"
            :class="['chat-conv-item', { active: to === conv.id }]"
            @click="selectConversation(conv.id)"
          >
            <div class="chat-conv-avatar">{{ conv.name.slice(0, 2).toUpperCase() }}</div>
            <div class="chat-conv-info">
              <strong>{{ conv.name }}</strong>
              <span>{{ conv.lastMessage.slice(0, 40) }}</span>
            </div>
          </button>
          <div v-if="!conversations.length" class="chat-conv-empty">
            <MessageSquare :size="24" />
            <span>{{ $t('chat.noConversations') || 'No conversations yet' }}</span>
          </div>
        </div>
      </aside>

      <!-- Message area -->
      <div class="chat-main">
        <div v-if="!to" class="chat-no-selection">
          <MessageSquare :size="56" />
          <h3>{{ $t('chat.selectConversation') || 'Select a conversation' }}</h3>
          <p>{{ $t('chat.selectMessage') || 'Choose a conversation from the list to start chatting' }}</p>
        </div>

        <template v-else>
          <div class="chat-main__header">
            <div class="chat-conv-avatar chat-conv-avatar--sm" style="cursor: pointer" @click="navigateTo('profile:' + to)">
              {{ (conversations.find(c => c.id === to)?.name || to).slice(0, 2).toUpperCase() }}
            </div>
            <div>
              <strong style="cursor: pointer" @click="navigateTo('profile:' + to)">
                {{ conversations.find(c => c.id === to)?.name || to }}
              </strong>
            </div>

            <div v-if="isClient" style="margin-left: auto; display: flex; align-items: center; gap: 0.5rem;">
              <!-- Payment inline -->
              <div v-if="showPaymentModal" style="display: flex; align-items: center; gap: 0.5rem; background: var(--surface); padding: 0.5rem; border-radius: var(--radius-md); border: 1px solid var(--border);">
                <input type="number" v-model="paymentAmount" min="1" step="10" class="input" style="width: 80px; padding: 0.25rem 0.5rem;" />
                <button class="button button--primary" @click="payFreelancer" :disabled="isPaying" style="padding: 0.25rem 0.75rem;">
                  {{ isPaying ? '...' : $t('chat.send') || 'Send' }}
                </button>
                <button class="button button--ghost" @click="showPaymentModal = false" style="padding: 0.25rem 0.5rem;">{{ $t('common.cancel') }}</button>
              </div>
              <button v-if="!showPaymentModal && !showReviewModal" class="button button--secondary" @click="showPaymentModal = true">
                <CircleDollarSign :size="16" />
                {{ $t('chat.pay') || 'Pay' }}
              </button>

              <!-- Review inline -->
              <div v-if="showReviewModal" style="display: flex; align-items: center; gap: 0.5rem; background: var(--surface); padding: 0.5rem; border-radius: var(--radius-md); border: 1px solid var(--border);">
                <div style="display: flex; gap: 2px;">
                  <Star
                    v-for="s in 5" :key="s" :size="20"
                    :style="{ cursor: 'pointer', color: s <= (reviewHover || reviewRating) ? '#f59e0b' : '#d1d5db', fill: s <= (reviewHover || reviewRating) ? '#f59e0b' : 'none' }"
                    @click="reviewRating = s"
                    @mouseenter="reviewHover = s"
                    @mouseleave="reviewHover = 0"
                  />
                </div>
                <input type="text" v-model="reviewComment" class="input" :placeholder="$t('chat.reviewPlaceholder') || 'Comment...'" style="width: 120px; padding: 0.25rem 0.5rem;" />
                <button class="button button--primary" @click="submitReview" :disabled="isSubmittingReview || reviewRating < 1" style="padding: 0.25rem 0.75rem;">
                  {{ isSubmittingReview ? '...' : $t('chat.send') || 'Send' }}
                </button>
                <button class="button button--ghost" @click="showReviewModal = false" style="padding: 0.25rem 0.5rem;">{{ $t('common.cancel') }}</button>
              </div>
              <button v-if="!showReviewModal && !showPaymentModal" class="button button--ghost" @click="showReviewModal = true">
                <Star :size="16" />
                {{ $t('chat.review') || 'Review' }}
              </button>
            </div>
          </div>

          <div ref="messagesContainer" class="chat-messages">
            <div
              v-for="(m, i) in activeConversation"
              :key="i"
              :class="['chat-bubble-wrap', m.from === 'me' ? 'chat-bubble-wrap--sent' : 'chat-bubble-wrap--received']"
            >
              <div :class="['chat-bubble', m.from === 'me' ? 'chat-bubble--sent' : 'chat-bubble--received']">
                <p>{{ m.content }}</p>
                <span class="chat-bubble__time">{{ formatTime(m.time) }}</span>
              </div>
            </div>
            <div v-if="!activeConversation.length" class="chat-messages-empty">
              <p>{{ $t('chat.startTyping') || 'Start typing to begin the conversation' }}</p>
            </div>
          </div>

          <form class="chat-compose" @submit.prevent="send">
            <input
              v-model="content"
              :placeholder="$t('chat.messagePlaceholder') || 'Type a message...'"
              type="text"
              autocomplete="off"
            />
            <button class="button button--primary chat-send-btn" type="submit" :disabled="!content.trim()">
              <Send :size="16" />
            </button>
          </form>
        </template>
      </div>
    </div>
  </section>
</template>
