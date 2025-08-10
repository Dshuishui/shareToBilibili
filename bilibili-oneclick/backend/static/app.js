// backend/static/app.js
document.addEventListener('DOMContentLoaded', function () {
    const form = document.getElementById('upload-form');
    const progressDiv = document.getElementById('progress');
    const resultDiv = document.getElementById('result');
  
    form.addEventListener('submit', function (e) {
      e.preventDefault();
      resultDiv.textContent = '';
      progressDiv.textContent = '进度：0%';
  
      const formData = new FormData(form);
      const xhr = new XMLHttpRequest();
      xhr.open('POST', '/upload');
  
      xhr.upload.addEventListener('progress', function (evt) {
        if (evt.lengthComputable) {
          const percentComplete = Math.round((evt.loaded / evt.total) * 100);
          progressDiv.textContent = '进度：' + percentComplete + '%';
        }
      });
  
      xhr.onload = function () {
        if (xhr.status >= 200 && xhr.status < 300) {
          const resp = JSON.parse(xhr.responseText);
          resultDiv.innerHTML = `<div class="ok">上传成功 · task_id: <b>${resp.task_id}</b></div>`;
        } else {
          resultDiv.innerHTML = `<div class="err">上传失败：${xhr.responseText}</div>`;
        }
      };
  
      xhr.onerror = function () {
        resultDiv.innerHTML = `<div class="err">上传时发生网络错误</div>`;
      };
  
      xhr.send(formData);
    });
  });
  document.getElementById('upload-form').addEventListener('submit', async function(e) {
    e.preventDefault();
    let formData = new FormData(this);

    let res = await fetch('/upload', { method: 'POST', body: formData });
    let data = await res.json();

    if (data.success) {
        document.getElementById('result').innerHTML = `
            <h3>上传成功，请确认信息：</h3>
            <pre>${JSON.stringify(data.metadata, null, 2)}</pre>
        `;
    }
});

document.getElementById('save-cookie-btn').addEventListener('click', async function() {
    let cookie = document.getElementById('cookie-input').value;
    let formData = new FormData();
    formData.append('cookie', cookie);

    let res = await fetch('/save_cookie', { method: 'POST', body: formData });
    let data = await res.json();
    alert(data.message);
});
