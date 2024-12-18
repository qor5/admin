package presets

import (
	"fmt"

	h "github.com/theplant/htmlgo"
)

func LinkageFieldsController(field *FieldContext, vs ...string) h.HTMLComponent {
	vs = append(vs, field.FormKey)
	return h.Div().Attr("v-on-mounted", fmt.Sprintf(`()=>{
	    dash.__lingkageFields = dash.__lingkageFields??[];
	    dash.__currentValidateKeys = dash.__currentValidateKeys??[];
		dash.__lingkageFields.push(%v)
		if (!vars.__findLinkageFields){
			dash.__findLinkageFields = function findLinkageFields( x) {
    		const result = new Set();
    		dash.__lingkageFields.forEach(subArray => {
        	if (subArray.includes(x)) {
            subArray.forEach(value => {	
			if (value !== x) {
				result.add(value);
				dash.__currentValidateKeys.push(value)
                }
            });
        }
    });
}
}
	}`, h.JSONString(vs)))
}
