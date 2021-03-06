package router_test

import (
	"net/http"
	"net/http/httptest"

	"github.com/awslabs/aws-sam-local/router"
	"github.com/awslabs/goformation/cloudformation"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

var _ = Describe("Function", func() {

	Context("with a GoFormation AWS::Serverless::Function", func() {

		r := router.NewServerlessRouter(false)

		function := &cloudformation.AWSServerlessFunction{
			Runtime: "nodejs6.10",
			Events: map[string]cloudformation.AWSServerlessFunction_EventSource{
				"GetRequests": cloudformation.AWSServerlessFunction_EventSource{
					Type: "Api",
					Properties: &cloudformation.AWSServerlessFunction_Properties{
						ApiEvent: &cloudformation.AWSServerlessFunction_ApiEvent{
							Path:   "/get",
							Method: "get",
						},
					},
				},
				"PostRequests": cloudformation.AWSServerlessFunction_EventSource{
					Type: "Api",
					Properties: &cloudformation.AWSServerlessFunction_Properties{
						ApiEvent: &cloudformation.AWSServerlessFunction_ApiEvent{
							Path:   "/post",
							Method: "post",
						},
					},
				},
				"AnyRequests": cloudformation.AWSServerlessFunction_EventSource{
					Type: "Api",
					Properties: &cloudformation.AWSServerlessFunction_Properties{
						ApiEvent: &cloudformation.AWSServerlessFunction_ApiEvent{
							Path:   "/any",
							Method: "any",
						},
					},
				},
				"ProxyResource": cloudformation.AWSServerlessFunction_EventSource{
					Type: "Api",
					Properties: &cloudformation.AWSServerlessFunction_Properties{
						ApiEvent: &cloudformation.AWSServerlessFunction_ApiEvent{
							Path:   "/proxy/{proxy+}",
							Method: "any",
						},
					},
				},
			},
		}

		err := r.AddFunction(function, func(w http.ResponseWriter, e *router.Event) {
			w.WriteHeader(200)
			w.Write([]byte("ok"))
		})

		It("should add the function successfully", func() {
			Expect(err).To(BeNil())
		})

		mounts := r.Mounts()
		It("should find three API event sources", func() {
			Expect(mounts).To(HaveLen(4))
		})

		It("should have the correct values for an event with GET http method", func() {
			Expect(mounts).To(ContainElement(PointTo(MatchFields(IgnoreExtras, Fields{
				"Name":      Equal("GetRequests"),
				"Path":      Equal("/get"),
				"Method":    Equal("get"),
			}))))
		})

		It("should have the correct values for an event with POST http method", func() {
			Expect(mounts).To(ContainElement(PointTo(MatchFields(IgnoreExtras, Fields{
				"Name":      Equal("PostRequests"),
				"Path":      Equal("/post"),
				"Method":    Equal("post"),
			}))))
		})

		It("should have the correct values for an event with ANY http method", func() {
			Expect(mounts).To(ContainElement(PointTo(MatchFields(IgnoreExtras, Fields{
				"Name":      Equal("AnyRequests"),
				"Path":      Equal("/any"),
				"Method":    Equal("any"),
			}))))
		})

		It("should have the correct values for an event with catchAll path", func() {
			Expect(mounts).To(ContainElement(PointTo(MatchFields(IgnoreExtras, Fields{
				"Name":      Equal("ProxyResource"),
				"Path":      Equal("/proxy/{proxy+}"),
				"Method":    Equal("any"),
			}))))
		})

		It("should respond to HTTP requests on GET /get", func() {
			req, _ := http.NewRequest("GET", "/get", nil)
			rr := httptest.NewRecorder()
			r.Router().ServeHTTP(rr, req)
			Expect(rr.Code).To(Equal(http.StatusOK))
			Expect(rr.Body.String()).To(Equal("ok"))
		})

		It("should respond to HTTP requests on POST /post", func() {
			req, _ := http.NewRequest("POST", "/post", nil)
			rr := httptest.NewRecorder()
			r.Router().ServeHTTP(rr, req)
			Expect(rr.Code).To(Equal(http.StatusOK))
			Expect(rr.Body.String()).To(Equal("ok"))
		})

		It("should respond to HTTP requests on POST /post", func() {
			req, _ := http.NewRequest("POST", "/post", nil)
			rr := httptest.NewRecorder()
			r.Router().ServeHTTP(rr, req)
			Expect(rr.Code).To(Equal(http.StatusOK))
			Expect(rr.Body.String()).To(Equal("ok"))
		})

		It("should respond with a 404 to HTTP requests on an invalid path", func() {
			req, _ := http.NewRequest("GET", "/invalid", nil)
			rr := httptest.NewRecorder()
			r.Router().ServeHTTP(rr, req)
			Expect(rr.Code).To(Equal(http.StatusNotFound))
		})

		It("should respond to GET requests on any path in catch-all resource", func() {
			req, _ := http.NewRequest("GET", "/proxy/hello/world", nil)
			rr := httptest.NewRecorder()
			r.Router().ServeHTTP(rr, req)
			Expect(rr.Code).To(Equal(http.StatusOK))
		})

		It("should respond to GET requests on a single sub-resource", func() {
			req, _ := http.NewRequest("GET", "/proxy/hello", nil)
			rr := httptest.NewRecorder()
			r.Router().ServeHTTP(rr, req)
			Expect(rr.Code).To(Equal(http.StatusOK))
		})

		It("should not respond to base proxy path, only sub-resources", func() {
			req, _ := http.NewRequest("GET", "/proxy", nil)
			rr := httptest.NewRecorder()
			r.Router().ServeHTTP(rr, req)
			Expect(rr.Code).To(Equal(http.StatusNotFound))
		})

		methods := []string{"OPTIONS", "GET", "HEAD", "POST", "PUT", "DELETE", "PATCH"}
		for _, method := range methods {
			It("should respond to HTTP requests on "+method+" /any", func() {
				req, _ := http.NewRequest(method, "/any", nil)
				rr := httptest.NewRecorder()
				r.Router().ServeHTTP(rr, req)
				Expect(rr.Code).To(Equal(http.StatusOK))
				Expect(rr.Body.String()).To(Equal("ok"))
			})
		}

	})

	Context("with a GoFormation AWS::Serverless::Function and prefix matching", func() {

		r := router.NewServerlessRouter(true)

		function := &cloudformation.AWSServerlessFunction{
			Runtime: "nodejs6.10",
			Events: map[string]cloudformation.AWSServerlessFunction_EventSource{
				"AnyRequests": cloudformation.AWSServerlessFunction_EventSource{
					Type: "Api",
					Properties: &cloudformation.AWSServerlessFunction_Properties{
						ApiEvent: &cloudformation.AWSServerlessFunction_ApiEvent{
							Path:   "/any",
							Method: "any",
						},
					},
				},
			},
		}

		err := r.AddFunction(function, func(w http.ResponseWriter, e *router.Event) {
			w.WriteHeader(200)
			w.Write([]byte("ok"))
		})

		It("should add the function successfully", func() {
			Expect(err).To(BeNil())
		})

		mounts := r.Mounts()
		It("should find three API event sources", func() {
			Expect(mounts).To(HaveLen(1))
		})

		It("should have the correct values for an event with ANY http method", func() {
			Expect(mounts).To(ContainElement(PointTo(MatchFields(IgnoreExtras, Fields{
				"Name":   Equal("AnyRequests"),
				"Path":   Equal("/any"),
				"Method": Equal("any"),
			}))))
		})

		It("should respond with a 404 to HTTP requests on an invalid path", func() {
			req, _ := http.NewRequest("GET", "/invalid", nil)
			rr := httptest.NewRecorder()
			r.Router().ServeHTTP(rr, req)
			Expect(rr.Code).To(Equal(http.StatusNotFound))
		})

		methods := []string{"OPTIONS", "GET", "HEAD", "POST", "PUT", "DELETE", "PATCH"}
		for _, method := range methods {
			It("should respond to HTTP requests on "+method+" /any", func() {
				req, _ := http.NewRequest(method, "/any", nil)
				rr := httptest.NewRecorder()
				r.Router().ServeHTTP(rr, req)
				Expect(rr.Code).To(Equal(http.StatusOK))
			})

			It("should respond to HTTP requests on "+method+" /any/foo", func() {
				req, _ := http.NewRequest(method, "/any/foo", nil)
				rr := httptest.NewRecorder()
				r.Router().ServeHTTP(rr, req)
				Expect(rr.Code).To(Equal(http.StatusNotFound))
			})
		}

	})

	Context("with a GoFormation AWS::Serverless::Function that has no 'Api' event sources", func() {

		r := router.NewServerlessRouter(false)

		function := &cloudformation.AWSServerlessFunction{
			Runtime: "nodejs6.10",
		}

		err := r.AddFunction(function, func(w http.ResponseWriter, e *router.Event) {
			w.WriteHeader(200)
			w.Write([]byte("ok"))
		})

		It("should throw a ErrNoEventsFound error", func() {
			Expect(err).To(MatchError(router.ErrNoEventsFound))
		})

	})

})
