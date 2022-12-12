
###2022-12-06 下载Livego 直播服务代码，准备分析学习一下
####简单使用
    1.启动服务：执行 livego 二进制文件启动 livego 服务；
    
    2.访问 http://localhost:8090/control/get?room=movie 获取一个房间的 channelkey(channelkey用于推流，movie用于播放).
        
    3.推流: 通过RTMP协议推送视频流到地址 rtmp://localhost:1935/{appname}/{channelkey} (appname默认是live), 例如： 使用 ffmpeg -re -i demo.flv -c copy -f flv rtmp://localhost:1935/{appname}/{channelkey} 推流(下载demo flv);
    ffmpeg -re -i demo.flv -c copy -f flv rtmp://localhost:1935/live/rfBd56ti2SMtYvSgD5xAV0YU99zampta7Z7S575KLkIZ9PYk    
    ffmpeg -re -i outputall.flv -c copy -f flv rtmp://localhost:1935/live/ABC
    
    4.播放: 支持多种播放协议，播放地址如下:    
    RTMP:rtmp://localhost:1935/{appname}/movie
    FLV:http://127.0.0.1:7001/{appname}/movie.flv
    HLS:http://127.0.0.1:7002/{appname}/movie.m3u8
    
   FLV:http://127.0.0.1:7001/live/movie.flv


    5.ffmpeg视频转换命令
    这是转换的基本命令：ffmpeg -i xxx.mp4 -qscale 1 -ar 44100 output.flv 但不指定输出的编码格式，不一定能推成功
    -qscale 1 //1最高品质 ，255最低品质
    -vcodec h264 //h264视频编码  livego可以成功处理
    -acodec aac  //acc音频编码 livego可以成功处理
    -ar是指 码率  太高不支持 
     
     下面是我测试可以正常直播，的flv转换命令；
     ffmpeg -i demo.mp4 -qscale 1 -vcodec h264 -acodec aac -ar 44100 output.flv
     
    6. ffmpeg將多個視頻合并的命令
    ffmpeg -f concat -i files.txt -c copy output.flv
    其中files.txt文件內容為要合并的文件清單
        file 'output1.flv'
        file 'output2.flv'
        file 'output3.flv'
        file 'output4.flv'
        file 'output5.flv'
        
        


###2022-12-09
    控制臺API信息
    1.獲得直播清單    http://localhost:8090/stat/livestat 
    2.獲得靜態資源    http://localhost:8090/statics/
    3.獲得房間信息    http://localhost:8090/control/get?room=movie ,這里的movie是個參數，未來在客戶端調用時要用
    4.刪房間   http://127.0.0.1:8090/control/delete?room=ROOM_NAME

###2022-12-10
#####房間號的處理
    1.源碼  configure/channel.go中 GetKey是，從RoomKeysType可以看到作者是有準備從redis來存儲token的，不過還沒有實現；
        type RoomKeysType struct {
            redisCli   *redis.Client
            localCache *cache.Cache
        }
    2.   