package presets

import (
	"fmt"

	h "github.com/theplant/htmlgo"
)

func LinkageFieldsController(field *FieldContext, vs ...string) h.HTMLComponent {
	vs = append(vs, field.FormKey)
	return h.Div().Attr("v-on-mounted", fmt.Sprintf(`()=>{
	    vars.__lingkageFields = vars.__lingkageFields??[];
		const endKey = %q;
		vars.__lingkageFields.push(%v)
		if (!vars.__findLinkageFields){
			vars.__findLinkageFields = function findLinkageFields( x) {
    		const result = new Set();
    		vars.__lingkageFields.forEach(subArray => {
        	if (subArray.includes(x)) {
            subArray.forEach(value => {	
			if (value !== x) {
				result.add(value);
				vars.__currentValidateKeys.push(key+endKey)
                }
            });
        }
    });
}
}
	}`, ErrorMessagePostfix, h.JSONString(vs)))
}
