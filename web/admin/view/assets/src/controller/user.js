
layui.define(['admin', 'form', 'jscrypt'], function(exports){
  var $ = layui.$
  ,setter = layui.setter
  ,admin = layui.admin
  ,form = layui.form
  ,jscrypt = layui.jscrypt
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
    var username = $(this).val();
    if (username.length == 0 || $('#LAY-user-login').data("lastcheck") == username) {
      return
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
    })
  });

  //提交
  form.on('submit(LAY-user-login-submit)', function(obj) {
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
      console.log(resp);
    });

    // admin.req({
    //   url: './json/user/login.js' //实际使用请改成服务端真实接口
    //   ,data: obj.field
    //   ,done: function(res) {

    //     //请求成功后，写入 access_token
    //     layui.data(setter.tableName, {key: setter.request.tokenName, value: res.data.access_token});

    //     //登入成功的提示与跳转
    //     layer.msg('登入成功', {
    //       offset: '15px'
    //       ,icon: 1
    //       ,time: 1000
    //     }, function(){
    //       location.hash = search.redirect ? decodeURIComponent(search.redirect) : '/';
    //     });
    //   }
    // });
    
  });
  
  
  //对外暴露的接口
  exports('user', {});
});