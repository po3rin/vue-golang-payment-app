<template>
  <div class="hello">
    <h1>{{ item.Name }}</h1>
    <h2>{{ item.Description }}</h2>
    <h2>{{ item.Amount }}円</h2>
    <payjp-checkout
      api-key="pk_test_892373dd331d28fa5152438e"
      client-id="d3d774f50bb006c26bac19402f0140a7228f8522"
      text="カードを情報を入力して購入"
      submit-text="購入確定"
      name-placeholder="田中 太郎"
      v-on:created="onTokenCreated"
      v-on:failed="onTokenFailed">
    </payjp-checkout>
    <p>{{ message }}</p>
    <router-link to="/">HOMEへ</router-link>
  </div>
</template>

<script>
import axios from 'axios'
export default {
  name: 'ItemCard',
  data () {
    return {
      item: {},
      message: ''
    }
  },
  created () {
    axios.get(`http://localhost:8888/api/v1/items/${this.$route.params.id}`).then(res => {
      this.item = res.data
    })
  },
  beforeDestroy () {
    window.PayjpCheckout = null
  },
  methods: {
    onTokenCreated: function (res) {
      console.log(res.id)
      const data = {Token: res.id}
      axios.post(`http://localhost:8888/api/v1/charge/items/${this.$route.params.id}`, data).then(res => {
        this.message = '商品の購入が完了しました！'
      })
    },
    onTokenFailed: function (status, err) {
      console.log(status)
      console.log(err)
    }
  }
}
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped>
h1, h2 {
  font-weight: normal;
}
ul {
  list-style-type: none;
  padding: 0;
}
li {
  display: inline-block;
  margin: 0 10px;
}
a {
  color: #42b983;
}
</style>
