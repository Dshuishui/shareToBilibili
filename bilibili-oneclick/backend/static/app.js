// backend/static/app.js
document.addEventListener('DOMContentLoaded', function () {
    const form = document.getElementById('upload-form');
    const cookieInput = document.getElementById('cookie-input');
    const saveCookieBtn = document.getElementById('save-cookie-btn');
    const uploadedInfoDiv = document.getElementById('uploaded-info');
    const publishBtn = document.getElementById('publish-btn');
    const publishStatusDiv = document.getElementById('publish-status');
  
    let lastTaskId = null;
  
    form.addEventListener('submit', async function (e) {
      e.preventDefault();
      publishStatusDiv.textContent = '';
      const fd = new FormData(form);
  
      try {
        const resp = await fetch('/upload', { method: 'POST', body: fd });
        const data = await resp.json();
        if (data.success) {
          lastTaskId = data.task_id;
          uploadedInfoDiv.innerHTML = `<pre>${JSON.stringify(data.metadata, null, 2)}</pre>`;
          publishBtn.disabled = false;
          publishStatusDiv.innerHTML = `<div style="color:green">上传并保存成功 · task_id: ${data.task_id}</div>`;
        } else {
          uploadedInfoDiv.textContent = '上传失败';
        }
      } catch (err) {
        console.error(err);
        uploadedInfoDiv.textContent = '上传时发生异常，请查看控制台';
      }
    });
  
    saveCookieBtn.addEventListener('click', async function () {
      const cookie = cookieInput.value.trim();
      if (!cookie) {
        alert('请先粘贴 cookie 字符串');
        return;
      }
      const fd = new FormData();
      fd.append('cookie', cookie);
      try {
        const resp = await fetch('/save_cookie', { method: 'POST', body: fd });
        const data = await resp.json();
        alert(data.message || JSON.stringify(data));
      } catch (err) {
        console.error(err);
        alert('保存 cookie 发生错误');
      }
    });
  
    publishBtn.addEventListener('click', async function () {
      publishStatusDiv.textContent = '';
      if (!lastTaskId) {
        alert('请先上传视频并确认信息');
        return;
      }
      publishBtn.disabled = true;
      publishStatusDiv.innerHTML = '开始投稿，请耐心等待（上传过程可能较慢）...';
  
      try {
        const resp = await fetch(`/publish/${lastTaskId}`, { method: 'POST' });
        if (!resp.ok) {
          const err = await resp.json();
          publishStatusDiv.innerHTML = `<div style="color:red">投稿失败：${err.detail || JSON.stringify(err)}</div>`;
          publishBtn.disabled = false;
          return;
        }
        const data = await resp.json();
        publishStatusDiv.innerHTML = `<div style="color:green">投稿接口返回：<pre>${JSON.stringify(data, null, 2)}</pre></div>`;
      } catch (err) {
        console.error(err);
        publishStatusDiv.innerHTML = `<div style="color:red">投稿过程中出现异常（查看控制台）</div>`;
      } finally {
        publishBtn.disabled = false;
      }
    });
  });
  