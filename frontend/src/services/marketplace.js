const baseJobs = [
  {
    id: 'job-101',
    title: 'Create a modern landing page for a new online course',
    client: 'Northwind Academy',
    budget: 1200,
    deadline: '2026-06-04',
    status: 'open',
    skills: ['Web Design', 'Copywriting', 'SEO', 'Analytics'],
    description: 'Design a clear page that explains the offer, collects leads, and works well on mobile.',
    proposals: 8,
  },
  {
    id: 'job-102',
    title: 'Build a booking flow for a small fitness studio',
    client: 'Apex Studio',
    budget: 900,
    deadline: '2026-05-29',
    status: 'inProgress',
    skills: ['UX', 'Frontend', 'Payments', 'Testing'],
    description: 'Improve the appointment flow and make it easier for customers to reserve sessions.',
    proposals: 5,
  },
  {
    id: 'job-103',
    title: 'Prepare brand visuals for a product launch',
    client: 'Brightline Group',
    budget: 650,
    deadline: '2026-06-11',
    status: 'open',
    skills: ['Branding', 'Social Media', 'Figma'],
    description: 'Create campaign visuals for Instagram, presentation slides, and product announcements.',
    proposals: 3,
  },
];

const API_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8088/api';
const FALLBACK_API_URL = 'http://localhost:8080/api';

async function request(path, options = {}) {
  const response = await fetchWithFallback(path, options);

  if (!response.ok) {
    const error = await response.json().catch(() => ({ error: response.statusText }));
    throw new Error(error.error || 'Request failed');
  }

  return response.json();
}

function getAuthHeaders() {
  const token = localStorage.getItem('fm_token');
  if (token) {
    return { Authorization: `Bearer ${token}` };
  }
  return {};
}

async function fetchWithFallback(path, options) {
  const requestOptions = {
    headers: {
      'Content-Type': 'application/json',
      ...getAuthHeaders(),
      ...(options.headers || {}),
    },
    ...options,
  };

  try {
    return await fetch(`${API_URL}${path}`, requestOptions);
  } catch (error) {
    if (API_URL === FALLBACK_API_URL) {
      throw error;
    }
    return fetch(`${FALLBACK_API_URL}${path}`, requestOptions);
  }
}


export const marketplaceApi = {
  async getSummary() {
    try {
      return await request('/summary');
    } catch {
      return {
        activeJobs: 24,
        escrowBalance: 18450,
        proposals: 132,
        rating: 4.8,
      };
    }
  },

  async getJobs() {
    try {
      return await request('/jobs');
    } catch {
      return structuredClone(baseJobs);
    }
  },

  async createJob(payload) {
    return request('/jobs', {
      method: 'POST',
      body: JSON.stringify(payload),
    });
  },

  async applyToJob(jobId, payload = {}) {
    return request(`/jobs/${jobId}/apply`, {
      method: 'POST',
      body: JSON.stringify(payload),
    });
  },

  async getProfile() {
    try {
      const userId = localStorage.getItem('fm_user_id');
      const query = userId ? `?user_id=${userId}` : '';
      return await request(`/profile${query}`);
    } catch {
      return {
        name: 'Guest User',
        role: 'freelancer',
        rating: 0,
        completedJobs: 0,
        skills: [],
      };
    }
  },

  async updateSkills(userId, skills) {
    return request('/profile/skills', {
      method: 'PUT',
      body: JSON.stringify({ userId, skills }),
    });
  },

  async updateProfile(payload) {
    return request('/profile', {
      method: 'PUT',
      body: JSON.stringify(payload),
    });
  },

  async uploadAvatar(file) {
    const token = localStorage.getItem('fm_token');
    const form = new FormData();
    form.append('file', file);
    const res = await fetch((import.meta.env.VITE_API_BASE_URL || 'http://localhost:8088/api').replace(/\/api$/, '') + '/api/profile/photo', {
      method: 'POST',
      headers: token ? { Authorization: `Bearer ${token}` } : {},
      body: form,
    });
    if (!res.ok) throw new Error('upload failed');
    return res.json();
  },

  async getPayments() {
    try {
      const userId = localStorage.getItem('fm_user_id');
      const query = userId ? `?user_id=${userId}` : '';
      return await request(`/payments${query}`);
    } catch {
      return {
        available: 4200,
        escrowed: 3100,
        history: [
          { id: 'tx-001', type: 'Deposit', amount: 2000, status: 'completed', date: '2026-05-10' },
          { id: 'tx-002', type: 'CreateEscrow', amount: 900, status: 'pending', date: '2026-05-11' },
          { id: 'tx-003', type: 'ReleasePayment', amount: 1250, status: 'completed', date: '2026-05-12' },
        ],
      };
    }
  },

  async deposit(amount) {
    return request('/payments/deposit', {
      method: 'POST',
      body: JSON.stringify({ amount }),
    });
  },
};
