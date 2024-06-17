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
        <Home />
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
import { Vue, Component } from 'vue-property-decorator'
import Home from './views/Home.vue'

@Component({
  components: {
    Home,
  }
})
export default class App extends Vue {
  created() {
    this.$vuetify.theme.dark = true
  }
}
</script>
