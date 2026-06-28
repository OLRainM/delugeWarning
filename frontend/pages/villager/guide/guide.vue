<template>
	<view class="container">
		<view class="card">
			<view style="font-weight:bold;font-size:32rpx;margin-bottom:16rpx;">洪水避险指引</view>
			<view v-for="(item, index) in tips" :key="index"
				style="padding:16rpx 0;border-bottom:1rpx solid #f0f0f0;">
				{{index+1}}. {{item}}
			</view>
		</view>
	</view>
</template>

<script>
	import api from '@/utils/request';
	export default {
		data() {
			return { tips: [] };
		},
		onLoad() {
			api.get('/api/v1/guides', { disaster_type: 'flood' }).then((res) => {
				this.tips = res.tips || [];
			});
		}
	};
</script>
