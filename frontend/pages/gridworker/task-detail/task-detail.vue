<template>
	<view class="container">
		<view class="card">
			<view style="font-weight:bold;">任务 #{{id}}</view>
			<view class="muted" style="margin-top:8rpx;">先点击"确认接收"，处置完成后填写说明并上传现场照片/视频。</view>
		</view>

		<button class="btn-line" @tap="confirm">确认接收任务</button>

		<view class="card" style="margin-top:20rpx;">
			<view style="font-weight:bold;margin-bottom:12rpx;">处置说明</view>
			<textarea style="width:100%;height:180rpx;" placeholder="请填写现场处置情况..."
				@input="onRemark" :value="remark"></textarea>
		</view>

		<view class="card" @tap="chooseMedia">
			<view style="color:#1f6feb;">+ 上传现场照片/视频（{{attachments.length}}）</view>
		</view>

		<button class="btn-primary" @tap="submit">提交处置结果</button>
	</view>
</template>

<script>
	import api from '@/utils/request';
	export default {
		data() {
			return { id: 0, remark: '', attachments: [] };
		},
		onLoad(q) { this.id = Number(q.id); },
		methods: {
			confirm() {
				api.post('/api/v1/tasks/' + this.id + '/confirm', {}).then(() => {
					uni.showToast({ title: '已确认接收' });
				});
			},
			onRemark(e) { this.remark = e.detail.value; },
			chooseMedia() {
				uni.chooseImage({
					count: 3,
					success: (res) => {
						res.tempFilePaths.forEach((p) => this.uploadOne(p, 'image'));
					}
				});
			},
			uploadOne(filePath, type) {
				const key = 'task/' + this.id + '/' + Date.now() + '.dat';
				api.post('/api/v1/uploads/presign', { key }).then((r) => {
					// 真机：uni.uploadFile 到 r.upload_url；此处直接登记访问地址
					this.attachments = this.attachments.concat([{ type, cos_key: key, url: r.access_url }]);
					uni.showToast({ title: '已添加附件' });
				});
			},
			submit() {
				api.post('/api/v1/tasks/' + this.id + '/handle', {
					remark: this.remark, attachments: this.attachments
				}).then(() => {
					uni.showToast({ title: '处置已提交' });
					setTimeout(() => uni.navigateBack(), 800);
				});
			}
		}
	};
</script>
