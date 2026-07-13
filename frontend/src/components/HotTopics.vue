<script setup lang="ts">
import {onBeforeMount, onUnmounted, ref} from 'vue'
import {HotTopic, OpenURL} from "../../wailsjs/go/main/App";
const list  = ref([])
const task =ref()

onBeforeMount(async () => {
  list.value = await HotTopic(10)
  task.value = setInterval(async ()=>{
    list.value = await HotTopic(10)
  }, 1000*10)
})
onUnmounted(()=>{
  clearInterval(task.value)
})

function showPage(htid) {
  if (!htid) return
  OpenURL(`https://gubatopic.eastmoney.com/topic_v3.html?htid=${encodeURIComponent(htid)}`)
}
</script>

<template>
  <n-list bordered hoverable clickable>
<!--    <template #header>-->
<!--      股吧热门-->
<!--    </template>-->
    <n-list-item v-for="(item, index) in list" :key="index">
        <n-thing :title="item.nickname" :description="item.desc" :description-style="'font-size: 14px;'"  @click="showPage(item.htid)">
          <template v-if="item.squareImg" #avatar>
            <n-avatar :src="item.squareImg" :size="60">
            </n-avatar>
          </template>
          <template v-if="item.stock_list" #footer>
            <n-flex>
              <n-tag type="info" v-for="(v, i) in item.stock_list" :bordered="false" size="small">
                {{v.name}}
              </n-tag>
            </n-flex>
          </template>
          <template v-if="item.clickNumber" #header-extra>
            <n-flex>
            <n-button secondary  type="warning" size="tiny">讨论数：<n-number-animation
                show-separator
                :from="0"
                :to="item.postNumber"
            />
            </n-button >
            <n-tag :bordered="false" type="warning" size="small">浏览量：<n-number-animation
                  show-separator
                  :from="0"
                  :to="item.clickNumber"
              />
            </n-tag>
            </n-flex>
          </template>
        </n-thing>
    </n-list-item>
  </n-list>
</template>

<style scoped>

</style>
