module.exports = {
  transpileDependencies: [
    'vuetify'
  ],
  devServer: {
    disableHostCheck: true,
    proxy: process.env.API_PROXY ? {
      '^/api': {
        target: process.env.API_PROXY,
      },
    } : undefined,
  }
}
