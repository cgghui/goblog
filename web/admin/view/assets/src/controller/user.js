
layui.define(['form', 'jsencrypt'], function(exports){
  var $ = layui.$
  ,setter = layui.setter
  ,form = layui.form
  ,jscrypt = layui.jsencrypt
  ,router = layui.router()
  ,search = router.search;

  form.render();

  form.verify({
    username: [/^[\S]{2,}$/, '请输入登录账号.'],
    password: [/^[\S]{6,}$/, '请输入登录密码.'],
  });

  $('body')

  .on('click', '#LAY-user-get-vercode', function() {
    var username = $('#LAY-user-login-username').val()
    if (!username) {
      return
    }
    var token = $('#captcha-token').val()
    if (!token) {
      return
    }
    $.get(setter.apiurl + 'auth/load_captcha?username=' + username + '&token=' + token + '&t='+ new Date().getTime()).done(function (resp) {
      $('#LAY-user-get-vercode').attr('src',resp.data.image);
    });
  })

  .on('blur', '#LAY-user-login-username', function () {
    checkuser($(this).val());
  });

  //提交
  form.on('submit(LAY-user-login-submit)', function(obj) {
    checkuser($('#LAY-user-login-username').val(), function () {
      var resp = $('#LAY-user-login').data();
      if (typeof(resp.code) == 'undefined') {
        layer.msg('请输入账号或密码', {icon: 2,time: 3000});
        return
      }
      if (resp.code == 11001) {
        layer.msg('登录账号或登录密码不正确', {icon: 2,time: 3000});
        return
      }
      if (resp.data.locked) {
        layer.msg('该账号存在异常行为，无法登录', {icon: 2,time: 3000});
        return
      }
    
      jscrypt.setPublicKey(resp.data.pubkey);
      obj.field.password = jscrypt.encrypt(obj.field.password);
  
      $.post(setter.apiurl + 'auth/passport', obj.field).done(function (resp) {
        if (resp.code == 11003) {
          if (resp.data.captcha_open) {
            $('#captcha-switch').show();
            $('#captcha-token').val(resp.data.captcha_token);
            $('#LAY-user-get-vercode').attr('src', resp.data.captcha_image);
          }
          return;
        }
        if (resp.code === 0 && resp.msg === "success") {
          layui.data(
            setter.tableName,
            {key: setter.request.tokenName, value: resp.data.access_token}
          );
          layer.msg('登入成功', {
            offset: '15px'
            ,icon: 1
            ,time: 1000
          }, function(){
            location.hash = search.redirect ? decodeURIComponent(search.redirect) : '/';
          });
        }
      });
    })
  });

  function checkuser (username, callfunc) {
    if (username.length == 0) {
      return;
    }
    if ($('#LAY-user-login').data("lastcheck") == username) {
      (typeof callfunc === "function") && callfunc();
      return;
    }
    $('#LAY-user-login').data({});
    $('#captcha-switch').hide();
    $('#captcha-token').val('');
    $('#LAY-user-get-vercode').attr('src', '');
    $('#rsa-pubkey').val('');
    $.get(setter.apiurl + 'auth/check?username=' + username).done(function (resp) {
      $('#LAY-user-login').data(resp);
      $('#LAY-user-login').data("lastcheck", username);
      if (resp.code != 0 || resp.data.locked) {
        return
      }
      if (resp.data.captcha_is_open) {
        $('#captcha-switch').show();
        $('#captcha-token').val(resp.data.captcha.token);
        $('#LAY-user-get-vercode').attr('src', resp.data.captcha.image);
      }
      $('#rsa-pubkey').val(resp.data.pubkey);
      (typeof callfunc === "function") && callfunc();
      $('#LAY-user-login').data({});
    })
  }

  //对外暴露的接口
  exports('user', {});
});