package controllers

import (
	"net/http"

	"github.com/labstack/echo/v5"
)

type Tracking struct{}

func NewTracking() Tracking {
	return Tracking{}
}

const trackingScript = `(function(){
  var d=document,s=d.currentScript,id=s.dataset.websiteId;
  var ep=new URL('/api/collect',s.src).href;
  function send(t,extra){
    var data={website_id:id,type:t,url:location.pathname+location.search,referrer:document.referrer,screen_width:window.innerWidth,language:navigator.language};
    if(extra){for(var k in extra){data[k]=extra[k]}}
    try{navigator.sendBeacon(ep,new Blob([JSON.stringify(data)],{type:'application/json'}))}catch(e){
      var x=new XMLHttpRequest();x.open('POST',ep);x.setRequestHeader('Content-Type','application/json');x.send(JSON.stringify(data));
    }
  }
  send('pageview');
  window.palantir={track:function(n,d){send('event',{event_name:n,event_data:d})}};
})();`

func (t Tracking) Script(etx *echo.Context) error {
	etx.Response().Header().Set("Content-Type", "application/javascript")
	etx.Response().Header().Set("Cache-Control", "public, max-age=86400")
	return etx.String(http.StatusOK, trackingScript)
}
