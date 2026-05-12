<script setup>
import { Activity, Bell, Gauge } from 'lucide-vue-next';

const targets = [
  { name: 'gateway', url: 'host.docker.internal:8080/metrics', ready: false },
  { name: 'user-service', url: 'host.docker.internal:8081/metrics', ready: false },
  { name: 'job-service', url: 'host.docker.internal:8082/metrics', ready: false },
  { name: 'payment-service', url: 'host.docker.internal:8083/metrics', ready: false },
  { name: 'prometheus', url: 'prometheus:9090/metrics', ready: true },
];
</script>

<template>
  <section class="view-stack">
    <div class="section-heading">
      <div>
        <p class="eyebrow">Grafana + Prometheus</p>
        <h1>{{ $t('monitoring.title') }}</h1>
      </div>
      <p>{{ $t('monitoring.subtitle') }}</p>
    </div>

    <div class="monitoring-grid">
      <article class="monitoring-tile">
        <Gauge :size="22" />
        <h2>{{ $t('monitoring.grafana') }}</h2>
        <p>http://localhost:3001</p>
      </article>
      <article class="monitoring-tile">
        <Activity :size="22" />
        <h2>{{ $t('monitoring.prometheus') }}</h2>
        <p>http://localhost:9090</p>
      </article>
      <article class="monitoring-tile">
        <Bell :size="22" />
        <h2>{{ $t('monitoring.alerts') }}</h2>
        <p>gateway-down, high-error-rate</p>
      </article>
    </div>

    <section class="table-panel">
      <div class="section-title">
        <h2>{{ $t('monitoring.target') }}</h2>
        <span>{{ $t('monitoring.dashboard') }}: freelance-market-overview</span>
      </div>
      <table>
        <thead>
          <tr>
            <th>Job</th>
            <th>URL</th>
            <th>{{ $t('monitoring.status') }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="target in targets" :key="target.name">
            <td>{{ target.name }}</td>
            <td>{{ target.url }}</td>
            <td>{{ target.ready ? $t('monitoring.up') : $t('monitoring.waiting') }}</td>
          </tr>
        </tbody>
      </table>
    </section>
  </section>
</template>
