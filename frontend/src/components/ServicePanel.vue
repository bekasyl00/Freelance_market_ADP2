<script setup>
import { Activity, Database, Mail, RadioTower, Server, ShieldCheck } from 'lucide-vue-next';

const services = [
  { key: 'User Service', icon: Server, items: ['Register', 'Login', 'GetProfile', 'UpdateSkills'] },
  { key: 'Job Service', icon: Activity, items: ['CreateJob', 'ListJobs', 'ApplyToJob', 'GetJobDetails', 'CompleteJob'] },
  { key: 'Payment Service', icon: ShieldCheck, items: ['Deposit', 'CreateEscrow', 'ReleasePayment', 'GetHistory'] },
  { key: 'NATS Worker', icon: RadioTower, items: ['Proposal event', 'Payment event', 'Email queue'] },
  { key: 'Supabase Postgres', icon: Database, items: ['users', 'jobs', 'proposals', 'transactions'] },
  { key: 'Email SMTP', icon: Mail, items: ['Application notification', 'Payment notification'] },
];
</script>

<template>
  <section class="service-grid" aria-label="Microservice readiness">
    <article v-for="service in services" :key="service.key" class="service-tile">
      <div class="service-tile__head">
        <component :is="service.icon" :size="20" />
        <h3>{{ service.key }}</h3>
      </div>
      <ul>
        <li v-for="item in service.items" :key="item">{{ item }}</li>
      </ul>
    </article>
  </section>
</template>
