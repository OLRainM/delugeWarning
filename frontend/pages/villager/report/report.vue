<template>
	<view class="container">
		<view class="card">
			<view style="font-weight:bold;margin-bottom:16rpx;">隐患描述</view>
			<textarea style="width:100%;height:200rpx;" placeholder="请描述发现的隐患情况..."
				@input="onContent" :value="content"></textarea>
		</view>
		<view class="card" @tap="locate">
			<view v-if="located">已定位：{{lng}}, {{lat}}</view>
			<view v-else style="color:#1f6feb;">点击获取当前定位 📍</view>
		</view>
		<button class="btn-primary" @tap="submit">提交上报</button>
	</view>
</template>

<script>
	import api from '@/utils/request';
	export default {
		data() {
			return { content: '', lng: 0, lat: 0, located: false };
		},
		methods: {
			onContent(e) { this.content = e.detail.value; },
			locate() {
				uni.getLocation({
					type: 'gcj02',
					success: (res) => {
						this.lng = res.longitude;
						this.lat = res.latitude;
						this.located = true;
					},
					fail: () => uni.showToast({ title: '定位失败', icon: 'none' })
				});
			},
			submit() {
				if (!this.content) { uni.showToast({ title: '请填写隐患描述', icon: 'none' }); return; }
				api.post('/api/v1/reports', {
					content: this.content, lng: this.lng, lat: this.lat
				}).then(() => {
					uni.showToast({ title: '上报成功' });
					setTimeout(() => uni.navigateBack(), 800);
				});
			}
		}
	};
</script>
