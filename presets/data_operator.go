package presets

import (
	"cmp"

	"github.com/pkg/errors"
	"github.com/qor5/web/v3"
	"github.com/qor5/x/v3/i18n"
	"github.com/qor5/x/v3/i18nx"
	"github.com/qor5/x/v3/statusx"
	"github.com/samber/lo"
	"golang.org/x/text/language"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// DataOperatorWithGRPC returns a hook that handles gRPC-related concerns for
// DataOperator: outgoing metadata injection and error conversion.
type grpcWrapper struct {
	next DataOperator
}

func DataOperatorWithGRPC(next DataOperator) DataOperator {
	return &grpcWrapper{next: next}
}

func (w *grpcWrapper) Search(eventCtx *web.EventContext, params *SearchParams) (*SearchResult, error) {
	w.inject(eventCtx)
	res, err := w.next.Search(eventCtx, params)
	if err != nil {
		return nil, w.convert(err)
	}
	return res, nil
}

func (w *grpcWrapper) Fetch(obj any, id string, eventCtx *web.EventContext) (any, error) {
	w.inject(eventCtx)
	res, err := w.next.Fetch(obj, id, eventCtx)
	if err != nil {
		return nil, w.convert(err)
	}
	return res, nil
}

func (w *grpcWrapper) Save(obj any, id string, eventCtx *web.EventContext) error {
	w.inject(eventCtx)
	if err := w.next.Save(obj, id, eventCtx); err != nil {
		return w.convert(err)
	}
	return nil
}

func (w *grpcWrapper) Delete(obj any, id string, eventCtx *web.EventContext) error {
	w.inject(eventCtx)
	if err := w.next.Delete(obj, id, eventCtx); err != nil {
		return w.convert(err)
	}
	return nil
}

func (w *grpcWrapper) inject(eventCtx *web.EventContext) {
	ctx := eventCtx.R.Context()
	ctx = metadata.AppendToOutgoingContext(
		ctx,
		i18nx.HeaderSelectedLanguage,
		i18n.LanguageTagFromContext(ctx, language.English).String(),
	)
	eventCtx.R = eventCtx.R.WithContext(ctx)
}

func (w *grpcWrapper) convert(err error) error {
	if err == nil {
		return nil
	}

	st, ok := status.FromError(err)
	if !ok {
		return err
	}

	var vErr web.ValidationErrors

	details := st.Details()

	badRequest := statusx.ExtractDetail[*errdetails.BadRequest](details)
	if badRequest != nil {
		for _, violation := range badRequest.GetFieldViolations() {
			vErr.FieldError(
				statusx.FormatField(violation.GetField(), lo.PascalCase),
				cmp.Or(violation.GetLocalizedMessage().GetMessage(), violation.GetDescription()),
			)
		}
		return errors.WithStack(&vErr)
	}

	localized := statusx.ExtractDetail[*errdetails.LocalizedMessage](details)
	if localized != nil {
		vErr.GlobalError(cmp.Or(localized.GetMessage(), st.Message()))
	} else {
		vErr.GlobalError(st.Message())
	}

	return errors.WithStack(&vErr)
}
