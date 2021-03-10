<template>
  <v-app>
    <v-app-bar
      app
      color="indigo"
      dark
    >
      <div>
        <a href="https://github.com/m1k1o/neko-rooms" style="text-decoration: none; color:white">neko-rooms</a>
      </div>

      <v-spacer></v-spacer>

      <div>
        by <a href="https://m1k1o.net" style="text-decoration: none; color:white">m1k1o</a>
      </div>
    </v-app-bar>

    <v-main>
      <v-container
        class="fill-height"
        fluid
      >
        <router-view />
      </v-container>
    </v-main>

    <v-footer app>
      <span>&copy; {{ new Date().getFullYear() }}</span>
      <v-spacer></v-spacer>
      <div>
        based on <a href="https://github.com/m1k1o/neko" style="text-decoration: none; color:white">n.eko</a>
      </div>
    </v-footer>
  </v-app>
</template>

<script lang="ts">
import Vue from 'vue'

export default Vue.extend({
  name: 'App',

  data: () => ({
    //
  }),

  created () {
    this.$vuetify.theme.dark = true
  }
})
</script>
