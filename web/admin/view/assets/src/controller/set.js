layui.define(['form', 'upload'], function(exports){
  var $ = layui.$
  ,layer = layui.layer
  ,laytpl = layui.laytpl
  ,setter = layui.setter
  ,view = layui.view
  ,admin = layui.admin
  ,form = layui.form
  ,upload = layui.upload;

  var $body = $('body');

  form.render();

  //自定义验证
  form.verify({
    username: function(value) {
      if(/^[a-zA-Z0-9_]+$/.test(value) == false) {
        return '用户名由字母、数字以及——组成';
      }
    }
    ,nickname: function(value) {
      if(/^[a-zA-Z0-9_\u4e00-\u9fa5]+$/.test(value) == false) {
        return '昵称由汉字、字母、数字以及_组成';
      }
    }

    //我们既支持上述函数式的方式，也支持下述数组的形式
    //数组的两个值分别代表：[正则匹配、匹配不符时的提示文字]
    ,pass: [
      /^[\S]{6,12}$/
      ,'密码必须6到12位，且不能出现空格'
    ]

    //确认密码
    ,repass: function(value){
      if(value !== $('#LAY_password').val()){
        return '两次密码输入不一致';
      }
    }
  });

  //网站设置
  form.on('submit(set_website)', function(obj){
    layer.msg(JSON.stringify(obj.field));

    //提交修改
    /*
    admin.req({
      url: ''
      ,data: obj.field
      ,success: function(){

      }
    });
    */
    return false;
  });

  //邮件服务
  form.on('submit(set_system_email)', function(obj){
    layer.msg(JSON.stringify(obj.field));

    //提交修改
    /*
    admin.req({
      url: ''
      ,data: obj.field
      ,success: function(){

      }
    });
    */
    return false;
  });


  //设置我的资料
  form.on('submit(setmyinfo)', function(obj){
    admin.req({
      method: "post",
      url: setter.apiurl + "auth/userinfo_update"
      ,data: obj.field
      ,success: function() {

      }
    });
  });



  //设置密码
  form.on('submit(setmypass)', function(obj){
    layer.msg(JSON.stringify(obj.field));

    //提交修改
    /*
    admin.req({
      url: ''
      ,data: obj.field
      ,success: function(){

      }
    });
    */
    return false;
  });

  //对外暴露的接口
  exports('set', {});
});