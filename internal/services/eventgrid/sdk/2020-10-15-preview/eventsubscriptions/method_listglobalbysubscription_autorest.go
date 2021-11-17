package eventsubscriptions

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure"
)

type ListGlobalBySubscriptionResponse struct {
	HttpResponse *http.Response
	Model        *[]EventSubscription

	nextLink     *string
	nextPageFunc func(ctx context.Context, nextLink string) (ListGlobalBySubscriptionResponse, error)
}

type ListGlobalBySubscriptionCompleteResult struct {
	Items []EventSubscription
}

func (r ListGlobalBySubscriptionResponse) HasMore() bool {
	return r.nextLink != nil
}

func (r ListGlobalBySubscriptionResponse) LoadMore(ctx context.Context) (resp ListGlobalBySubscriptionResponse, err error) {
	if !r.HasMore() {
		err = fmt.Errorf("no more pages returned")
		return
	}
	return r.nextPageFunc(ctx, *r.nextLink)
}

type ListGlobalBySubscriptionOptions struct {
	Filter *string
	Top    *int64
}

func DefaultListGlobalBySubscriptionOptions() ListGlobalBySubscriptionOptions {
	return ListGlobalBySubscriptionOptions{}
}

func (o ListGlobalBySubscriptionOptions) toQueryString() map[string]interface{} {
	out := make(map[string]interface{})

	if o.Filter != nil {
		out["$filter"] = *o.Filter
	}

	if o.Top != nil {
		out["$top"] = *o.Top
	}

	return out
}

// ListGlobalBySubscription ...
func (c EventSubscriptionsClient) ListGlobalBySubscription(ctx context.Context, id SubscriptionId, options ListGlobalBySubscriptionOptions) (resp ListGlobalBySubscriptionResponse, err error) {
	req, err := c.preparerForListGlobalBySubscription(ctx, id, options)
	if err != nil {
		err = autorest.NewErrorWithError(err, "eventsubscriptions.EventSubscriptionsClient", "ListGlobalBySubscription", nil, "Failure preparing request")
		return
	}

	resp.HttpResponse, err = c.Client.Send(req, azure.DoRetryWithRegistration(c.Client))
	if err != nil {
		err = autorest.NewErrorWithError(err, "eventsubscriptions.EventSubscriptionsClient", "ListGlobalBySubscription", resp.HttpResponse, "Failure sending request")
		return
	}

	resp, err = c.responderForListGlobalBySubscription(resp.HttpResponse)
	if err != nil {
		err = autorest.NewErrorWithError(err, "eventsubscriptions.EventSubscriptionsClient", "ListGlobalBySubscription", resp.HttpResponse, "Failure responding to request")
		return
	}
	return
}

// ListGlobalBySubscriptionComplete retrieves all of the results into a single object
func (c EventSubscriptionsClient) ListGlobalBySubscriptionComplete(ctx context.Context, id SubscriptionId, options ListGlobalBySubscriptionOptions) (ListGlobalBySubscriptionCompleteResult, error) {
	return c.ListGlobalBySubscriptionCompleteMatchingPredicate(ctx, id, options, EventSubscriptionPredicate{})
}

// ListGlobalBySubscriptionCompleteMatchingPredicate retrieves all of the results and then applied the predicate
func (c EventSubscriptionsClient) ListGlobalBySubscriptionCompleteMatchingPredicate(ctx context.Context, id SubscriptionId, options ListGlobalBySubscriptionOptions, predicate EventSubscriptionPredicate) (resp ListGlobalBySubscriptionCompleteResult, err error) {
	items := make([]EventSubscription, 0)

	page, err := c.ListGlobalBySubscription(ctx, id, options)
	if err != nil {
		err = fmt.Errorf("loading the initial page: %+v", err)
		return
	}
	if page.Model != nil {
		for _, v := range *page.Model {
			if predicate.Matches(v) {
				items = append(items, v)
			}
		}
	}

	for page.HasMore() {
		page, err = page.LoadMore(ctx)
		if err != nil {
			err = fmt.Errorf("loading the next page: %+v", err)
			return
		}

		if page.Model != nil {
			for _, v := range *page.Model {
				if predicate.Matches(v) {
					items = append(items, v)
				}
			}
		}
	}

	out := ListGlobalBySubscriptionCompleteResult{
		Items: items,
	}
	return out, nil
}

// preparerForListGlobalBySubscription prepares the ListGlobalBySubscription request.
func (c EventSubscriptionsClient) preparerForListGlobalBySubscription(ctx context.Context, id SubscriptionId, options ListGlobalBySubscriptionOptions) (*http.Request, error) {
	queryParameters := map[string]interface{}{
		"api-version": defaultApiVersion,
	}

	for k, v := range options.toQueryString() {
		queryParameters[k] = autorest.Encode("query", v)
	}

	preparer := autorest.CreatePreparer(
		autorest.AsContentType("application/json; charset=utf-8"),
		autorest.AsGet(),
		autorest.WithBaseURL(c.baseUri),
		autorest.WithPath(fmt.Sprintf("%s/providers/Microsoft.EventGrid/eventSubscriptions", id.ID())),
		autorest.WithQueryParameters(queryParameters))
	return preparer.Prepare((&http.Request{}).WithContext(ctx))
}

// preparerForListGlobalBySubscriptionWithNextLink prepares the ListGlobalBySubscription request with the given nextLink token.
func (c EventSubscriptionsClient) preparerForListGlobalBySubscriptionWithNextLink(ctx context.Context, nextLink string) (*http.Request, error) {
	uri, err := url.Parse(nextLink)
	if err != nil {
		return nil, fmt.Errorf("parsing nextLink %q: %+v", nextLink, err)
	}
	queryParameters := map[string]interface{}{}
	for k, v := range uri.Query() {
		if len(v) == 0 {
			continue
		}
		val := v[0]
		val = autorest.Encode("query", val)
		queryParameters[k] = val
	}

	preparer := autorest.CreatePreparer(
		autorest.AsContentType("application/json; charset=utf-8"),
		autorest.AsGet(),
		autorest.WithBaseURL(c.baseUri),
		autorest.WithPath(uri.Path),
		autorest.WithQueryParameters(queryParameters))
	return preparer.Prepare((&http.Request{}).WithContext(ctx))
}

// responderForListGlobalBySubscription handles the response to the ListGlobalBySubscription request. The method always
// closes the http.Response Body.
func (c EventSubscriptionsClient) responderForListGlobalBySubscription(resp *http.Response) (result ListGlobalBySubscriptionResponse, err error) {
	type page struct {
		Values   []EventSubscription `json:"value"`
		NextLink *string             `json:"nextLink"`
	}
	var respObj page
	err = autorest.Respond(
		resp,
		azure.WithErrorUnlessStatusCode(http.StatusOK),
		autorest.ByUnmarshallingJSON(&respObj),
		autorest.ByClosing())
	result.HttpResponse = resp
	result.Model = &respObj.Values
	result.nextLink = respObj.NextLink
	if respObj.NextLink != nil {
		result.nextPageFunc = func(ctx context.Context, nextLink string) (result ListGlobalBySubscriptionResponse, err error) {
			req, err := c.preparerForListGlobalBySubscriptionWithNextLink(ctx, nextLink)
			if err != nil {
				err = autorest.NewErrorWithError(err, "eventsubscriptions.EventSubscriptionsClient", "ListGlobalBySubscription", nil, "Failure preparing request")
				return
			}

			result.HttpResponse, err = c.Client.Send(req, azure.DoRetryWithRegistration(c.Client))
			if err != nil {
				err = autorest.NewErrorWithError(err, "eventsubscriptions.EventSubscriptionsClient", "ListGlobalBySubscription", result.HttpResponse, "Failure sending request")
				return
			}

			result, err = c.responderForListGlobalBySubscription(result.HttpResponse)
			if err != nil {
				err = autorest.NewErrorWithError(err, "eventsubscriptions.EventSubscriptionsClient", "ListGlobalBySubscription", result.HttpResponse, "Failure responding to request")
				return
			}

			return
		}
	}
	return
}