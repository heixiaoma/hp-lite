<template>
  <div id="qr">
    <canvas id="canvas"></canvas>
    <div v-if="showText">{{text}}</div>
  </div>
</template>

<script>
import QRCode from 'qrcode';

export default {
  props: {
    text: {
      type: String,
      required: true
    } ,
    showText: {
      type: Boolean,
      default:true,
      required: false
    }
  },
  mounted() {
    let canvas = document.getElementById('canvas')
    QRCode.toCanvas(canvas, this.text, function (error) {
      if (error) console.error(error)
      console.log('success!');
    })
  }
}
</script>
<style>
#qr {
  text-align: center;
}
</style>
