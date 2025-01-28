<script setup>
import { RouterView } from 'vue-router'
import { ref, computed } from 'vue'

const currentUsername = ref(localStorage.getItem('username'))

const isLoggedIn = computed(() => {
  return localStorage.getItem('sessionId') !== null
})
</script>

<template>
  <router-view/>
</template>

<style>
/* Keep any global styles you need */
</style>
