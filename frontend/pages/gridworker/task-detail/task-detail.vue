<template>
	<view class="container">

		<!-- ===== 已完成：只读预览 ===== -->
		<view v-if="task.status === 'finished'">
			<view class="card" style="border-left:8rpx solid #52c41a;">
				<view style="display:flex;align-items:center;gap:12rpx;margin-bottom:12rpx;">
					<text class="tag" style="background:#52c41a;">✅ 已完成处置</text>
					<text class="muted">任务 #{{task.id}}</text>
				</view>
				<view style="font-size:28rpx;font-weight:bold;margin-bottom:8rpx;">处置说明</view>
				<view style="line-height:1.6;color:#333;">{{task.handle_remark || '（未填写说明）'}}</view>
				<view class="muted" style="margin-top:16rpx;">完成时间：{{finishedAt}}</view>
			</view>

			<view class="card" v-if="attachments.length > 0">
				<view style="font-weight:bold;margin-bottom:16rpx;">现场附件（{{attachments.length}}）</view>
				<view style="display:flex;flex-wrap:wrap;gap:16rpx;">
					<image v-for="att in attachments" :key="att.id"
						v-if="att.type==='image'" :src="att.url" mode="aspectFill"
						style="width:200rpx;height:200rpx;border-radius:8rpx;"
						@tap="preview(att.url)" />
					<view v-else style="width:200rpx;height:200rpx;background:#f0f0f0;border-radius:8rpx;
						display:flex;align-items:center;justify-content:center;color:#666;font-size:24rpx;">
						🎬 视频
					</view>
				</view>
			</view>
			<view class="card muted" v-else style="text-align:center;">暂无现场附件</view>

			<view class="card" v-if="alert.id">
				<view style="font-weight:bold;margin-bottom:8rpx;">关联预警</view>
				<view>{{alert.title}}</view>
				<view class="muted" style="margin-top:4rpx;">{{alert.content}}</view>
			</view>
		</view>

		<!-- ===== 待确认 / 处置中：操作表单 ===== -->
		<view v-else>
			<view class="card">
				<view style="font-weight:bold;">任务 #{{task.id || id}}</view>
				<view class="muted" style="margin-top:8rpx;">
					{{task.status==='pending'?'请先确认接收，再到现场处置并提交结果。':'请填写处置说明并上传现场照片/视频后提交。'}}
				</view>
			</view>

			<button class="btn-line" v-if="task.status==='pending' || !task.status" @tap="confirm">确认接收任务</button>

			<view class="card" style="margin-top:20rpx;">
				<view style="font-weight:bold;margin-bottom:12rpx;">处置说明</view>
				<textarea style="width:100%;height:180rpx;" placeholder="请填写现场处置情况..."
					@input="onRemark" :value="remark"></textarea>
			</view>

			<view class="card" @tap="chooseMedia">
				<view style="color:#1f6feb;">+ 上传现场照片/视频（{{attachments.length}}）</view>
			</view>

			<button class="btn-primary" v-if="task.status==='handling'" @tap="submit">提交处置结果</button>
		</view>
	</view>
</template>

<script>
	import api from '@/utils/request';
	import fmt from '@/utils/format';
	export default {
		data() {
			return { id: 0, task: {}, alert: {}, attachments: [], remark: '', finishedAt: '' };
		},
		onLoad(q) {
			this.id = Number(q.id);
			api.get('/api/v1/tasks', { status: '' }).then((res) => {
				const t = (res.items || []).find(x => x.id === this.id);
				if (!t) return;
				this.task = t;
				this.finishedAt = fmt.fmtTime(t.finished_at);
				if (t.status === 'finished') {
					api.get('/api/v1/alerts/' + t.alert_id).then(r => {
						this.alert = r.alert || {};
					}).catch(() => {});
				}
			});
		},
		methods: {
			confirm() {
				api.post('/api/v1/tasks/' + this.id + '/confirm', {}).then(() => {
					this.task = { ...this.task, status: 'handling' };
					uni.showToast({ title: '已确认接收' });
				});
			},
			onRemark(e) { this.remark = e.detail.value; },
			chooseMedia() {
				uni.chooseImage({ count: 3, success: (res) => {
					res.tempFilePaths.forEach(p => this.uploadOne(p, 'image'));
				}});
			},
			uploadOne(filePath, type) {
				const key = 'task/' + this.id + '/' + Date.now() + '.dat';
				api.post('/api/v1/uploads/presign', { key }).then(r => {
					this.attachments = this.attachments.concat([{ type, cos_key: key, url: r.access_url }]);
					uni.showToast({ title: '已添加' });
				});
			},
			submit() {
				api.post('/api/v1/tasks/' + this.id + '/handle', {
					remark: this.remark, attachments: this.attachments
				}).then(() => {
					uni.showToast({ title: '处置已提交' });
					setTimeout(() => uni.navigateBack(), 800);
				});
			},
			preview(url) { uni.previewImage({ urls: [url] }); }
		}
	};
</script>

