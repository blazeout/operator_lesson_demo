/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"github.com/kubebuilder-demo/controllers/utils"
	_ "github.com/kubebuilder-demo/controllers/utils"
	"google.golang.org/appengine/log"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	_ "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	_ "k8s.io/apimachinery/pkg/types"
	_ "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	_ "sigs.k8s.io/controller-runtime/pkg/log"

	ingressv1beta1 "github.com/kubebuilder-demo/api/v1beta1"
)

// AppReconciler reconciles a App object
type AppReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=ingress.baiding.tech,resources=apps,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=ingress.baiding.tech,resources=apps/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=ingress.baiding.tech,resources=apps/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the App object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.2/pkg/reconcile
func (r *AppReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	//todo
	// 先获取到App对象
	app := &ingressv1beta1.App{}
	err := r.Get(ctx, req.NamespacedName, app)
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	// 1. 创建Deployment对象
	deployment := utils.NewDeployment(app)
	err = ctrl.SetControllerReference(app, deployment, r.Scheme)
	if err != nil {
		return ctrl.Result{}, err
	}
	// 创建完成之后就查询是否存在deployment 如果不存在就创建, 如果存在就更新
	d := &v1.Deployment{}
	if err = r.Get(ctx, req.NamespacedName, d); err != nil {
		// 如果不存在就创建
		if errors.IsNotFound(err) {
			if err := r.Create(ctx, deployment); err != nil {
				log.Errorf(ctx, "Failed to create new Deployment: %v", err)
				return ctrl.Result{}, err
			}
		}
	} else {
		// 如果存在就更新
		if err = r.Update(ctx, deployment); err != nil {
			return ctrl.Result{}, err
		}
	}
	// 2. Service的处理
	service := utils.NewService(app)
	err = ctrl.SetControllerReference(app, service, r.Scheme)
	if err != nil {
		return ctrl.Result{}, err
	}
	// 查找指定的Service
	s := &corev1.Service{}
	if err := r.Get(ctx, types.NamespacedName{Namespace: app.Namespace, Name: app.Name}, s); err != nil {
		if errors.IsNotFound(err) && app.Spec.EnableService {
			if err := r.Create(ctx, service); err != nil {
				log.Errorf(ctx, "Failed to create new Service: %v", err)
				return ctrl.Result{}, err
			}
		} else if !errors.IsNotFound(err) && app.Spec.EnableService {
			// 如果错误不是NotFound 并且EnableService为true, 则需要重试
			return ctrl.Result{}, err
		}
	} else {
		if app.Spec.EnableService {
			if err := r.Update(ctx, service); err != nil {
				return ctrl.Result{}, err
			}
		} else {
			if err := r.Delete(ctx, service); err != nil {
				return ctrl.Result{}, err
			}
		}
	}

	// 3. ingress的处理
	//todo 使用admission webhook, 如果启动了ingress, 那么Service也启动
	//todo 默认为false
	if !app.Spec.EnableService {
		return ctrl.Result{}, nil
	}
	ingress := utils.NewIngress(app)
	err = ctrl.SetControllerReference(app, ingress, r.Scheme)
	if err != nil {
		if errors.IsNotFound(err) && app.Spec.EnableIngress {
			if err := r.Create(ctx, ingress); err != nil {
				log.Errorf(ctx, "Failed to create new Ingress: %v", err)
				return ctrl.Result{}, err
			}
		} else if !errors.IsNotFound(err) && app.Spec.EnableIngress {
			// 如果错误不是NotFound 并且EnableIngress为true, 则需要重试
			return ctrl.Result{}, err
		}
	} else {
		if app.Spec.EnableIngress {
			err := r.Update(ctx, ingress)
			if err != nil {
				return ctrl.Result{}, err
			}
		} else {
			if err := r.Delete(ctx, ingress); err != nil {
				return ctrl.Result{}, err
			}
		}
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *AppReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&ingressv1beta1.App{}).
		Owns(&v1.Deployment{}).
		Owns(&netv1.Ingress{}).
		Owns(&corev1.Service{}).
		Complete(r)
}
