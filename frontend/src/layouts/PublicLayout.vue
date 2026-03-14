<script setup lang="ts">
import { RouterView, useRoute } from "vue-router";

import Navbar from "../components/Navbar.vue";

const route = useRoute();

const links = [
  { to: "/", label: "首页" },
  { to: "/upload", label: "上传" },
];
</script>

<template>
  <div class="app-shell">
    <Navbar
      :items="links"
      :current-path="route.path"
      github-href="https://github.com/zzzzquan/OpenShare"
    />

    <main class="pt-16">
      <RouterView />
    </main>
  </div>
</template>
