<template>
  <div class="itemcard">
    <h1>{{ item.Name }}</h1>
    <h2>{{ item.Description }}</h2>
    <h2>{{ item.Amount }}円</h2>
  </div>
</template>

<script>
export default {
  name: 'ItemCard',
  props: [
    'item'
  ]
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
.itemcard {
  border: solid 1px gray;
}
</style>
