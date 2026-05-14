<script setup>
import { ref, onMounted, onBeforeUnmount, nextTick, computed } from 'vue';
import { Send, MessageSquare, Users, Search, AlertCircle, Wifi, WifiOff } from 'lucide-vue-next';

const ws = ref(null);
const messages = ref([]);
const to = ref('');
const content = ref('');
const error = ref('');
const connected = ref(false);
const searchQuery = ref('');
const messagesContainer = ref(null);
const currentUserId = ref(localStorage.getItem('fm_user_id') || '');

// Group messages by conversation partner
const conversations = computed(() => {
  const convMap = {};
  messages.value.forEach(m => {
    const partner = m.from === 'me' ? (m.to || to.value) : m.from;
    if (!partner) return;
    if (!convMap[partner]) {
      convMap[partner] = { id: partner, messages: [], lastMessage: '', lastTime: new Date() };
    }
    convMap[partner].messages.push(m);
    convMap[partner].lastMessage = m.content;
  });
  return Object.values(convMap);
});

const activeConversation = computed(() => {
  if (!to.value) return [];
  return messages.value.filter(m =>
    (m.from === 'me' && (m.to === to.value || !m.to)) ||
    m.from === to.value
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

onMounted(() => {
  const token = localStorage.getItem('fm_token');
  if (!token) { error.value = 'not_authenticated'; return; }

  const baseUrl = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8088/api';
  const apiHost = baseUrl.replace(/^https?:\/\//, '').replace(/\/api$/, '');
  const protocol = baseUrl.startsWith('https') ? 'wss' : 'ws';

  let wsUrl;
  if (import.meta.env.VITE_API_BASE_URL) {
    wsUrl = `${protocol}://${apiHost}/ws/chat?token=${token}`;
  } else {
    const host = location.host;
    const wsProtocol = location.protocol === 'https:' ? 'wss' : 'ws';
    wsUrl = `${wsProtocol}://${host}/ws/chat?token=${token}`;
  }

  try {
    ws.value = new WebSocket(wsUrl);
    ws.value.addEventListener('open', () => { connected.value = true; });
    ws.value.addEventListener('close', () => { connected.value = false; });
    ws.value.addEventListener('error', () => { connected.value = false; });
    ws.value.addEventListener('message', (e) => {
      try {
        const data = JSON.parse(e.data);
        addMessage({ from: data.from, content: data.content, time: new Date() });
      } catch (err) { /* ignore */ }
    });
  } catch {
    error.value = 'connection_failed';
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

        <!-- New conversation -->
        <div class="chat-new-conv">
          <div class="auth-field__input-wrap chat-user-input">
            <Search :size="16" class="auth-field__icon" />
            <input
              v-model="to"
              type="text"
              :placeholder="$t('chat.recipientPlaceholder') || 'Enter user ID...'"
            />
          </div>
        </div>

        <!-- Conversation list -->
        <div class="chat-conv-list">
          <button
            v-for="conv in conversations"
            :key="conv.id"
            :class="['chat-conv-item', { active: to === conv.id }]"
            @click="selectConversation(conv.id)"
          >
            <div class="chat-conv-avatar">{{ conv.id.slice(0, 2).toUpperCase() }}</div>
            <div class="chat-conv-info">
              <strong>{{ conv.id.slice(0, 8) }}…</strong>
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
          <p>{{ $t('chat.selectMessage') || 'Enter a user ID to start chatting' }}</p>
        </div>

        <template v-else>
          <div class="chat-main__header">
            <div class="chat-conv-avatar chat-conv-avatar--sm">{{ to.slice(0, 2).toUpperCase() }}</div>
            <div>
              <strong>{{ to.slice(0, 12) }}…</strong>
              <span class="chat-status chat-status--online">
                <Wifi :size="10" />
                {{ $t('common.online') || 'Online' }}
              </span>
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
