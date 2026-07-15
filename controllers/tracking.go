package controllers

import (
	"net/http"

	"palantir/router"
	"palantir/router/routes"

	"github.com/labstack/echo/v5"
)

type Tracking struct{}

func NewTracking() Tracking { return Tracking{} }

func (t Tracking) RegisterRoutes(r *router.Router) error {
	_, err := r.AddRoute(echo.Route{Method: http.MethodGet, Path: routes.TrackingScript.Path(), Name: routes.TrackingScript.Name(), Handler: t.Script})
	return err
}

const trackingScript = `(function(){
var d=document,s=d.currentScript,id=s&&s.dataset.websiteId;if(!id)return;
var ep=new URL('/api/collect',s.src).href;
function send(type,extra){var data={website_id:id,type:type,url:location.pathname+location.search,referrer:d.referrer,screen_width:innerWidth,language:navigator.language};
if(extra)Object.assign(data,extra);var body=JSON.stringify(data);
if(navigator.sendBeacon&&navigator.sendBeacon(ep,new Blob([body],{type:'application/json'})))return;
fetch(ep,{method:'POST',headers:{'Content-Type':'application/json'},body:body,keepalive:true}).catch(function(){});}
send('pageview');window.palantir={track:function(name,data){send('event',{event_name:name,event_data:data})}};
})();`

func (Tracking) Script(c *echo.Context) error {
	c.Response().Header().Set("Cache-Control", "public, max-age=86400")
	return c.Blob(http.StatusOK, "application/javascript; charset=utf-8", []byte(trackingScript))
}
