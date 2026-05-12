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

export const marketplaceApi = {
  async getSummary() {
    return {
      activeJobs: 24,
      escrowBalance: 18450,
      proposals: 132,
      rating: 4.8,
    };
  },

  async getJobs() {
    return structuredClone(baseJobs);
  },

  async getProfile() {
    return {
      name: 'Aruzhan Karimova',
      role: 'freelancer',
      rating: 4.8,
      completedJobs: 37,
      skills: ['Web Design', 'Frontend', 'Branding', 'SEO', 'Payments'],
    };
  },

  async getPayments() {
    return {
      available: 4200,
      escrowed: 3100,
      history: [
        { id: 'tx-001', type: 'Deposit', amount: 2000, status: 'completed', date: '2026-05-10' },
        { id: 'tx-002', type: 'CreateEscrow', amount: 900, status: 'pending', date: '2026-05-11' },
        { id: 'tx-003', type: 'ReleasePayment', amount: 1250, status: 'completed', date: '2026-05-12' },
      ],
    };
  },
};
