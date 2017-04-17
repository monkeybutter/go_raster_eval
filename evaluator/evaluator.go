package evaluator

import (
	"../ast"
	"../object"
	"../raster"
	"fmt"
)

var (
	NULL  = &object.Null{}
	TRUE  = &object.Boolean{Value: true}
	FALSE = &object.Boolean{Value: false}
)

func Eval(node ast.Node, env *object.Environment) object.Object {
	switch node := node.(type) {

	// Statements
	case *ast.Program:
		return evalProgram(node, env)

	case *ast.BlockStatement:
		return evalBlockStatement(node, env)

	case *ast.ExpressionStatement:
		return Eval(node.Expression, env)

	// Expressions
	case *ast.NumberLiteral:
		return &object.Number{Value: node.Value}

	case *ast.Boolean:
		return nativeBoolToBooleanObject(node.Value)

	case *ast.PrefixExpression:
		right := Eval(node.Right, env)
		if isError(right) {
			return right
		}
		return evalPrefixExpression(node.Operator, right)

	case *ast.InfixExpression:
		left := Eval(node.Left, env)
		if isError(left) {
			return left
		}

		right := Eval(node.Right, env)
		if isError(right) {
			return right
		}

		return evalInfixExpression(node.Operator, left, right)

	case *ast.Identifier:
		return evalIdentifier(node, env)
	}

	return nil
}

func evalProgram(program *ast.Program, env *object.Environment) object.Object {
	var result object.Object

	for _, statement := range program.Statements {
		result = Eval(statement, env)

		switch result := result.(type) {
		case *object.ReturnValue:
			return result.Value
		case *object.Error:
			return result
		}
	}

	return result
}

func evalBlockStatement(block *ast.BlockStatement, env *object.Environment) object.Object {
	var result object.Object

	for _, statement := range block.Statements {
		result = Eval(statement, env)

		if result != nil {
			rt := result.Type()
			if rt == object.RETURN_VALUE_OBJ || rt == object.ERROR_OBJ {
				return result
			}
		}
	}

	return result
}

func nativeBoolToBooleanObject(input bool) *object.Boolean {
	if input {
		return TRUE
	}
	return FALSE
}

func evalPrefixExpression(operator string, right object.Object) object.Object {
	switch operator {
	case "!":
		return evalBangOperatorExpression(right)
	case "-":
		return evalMinusPrefixOperatorExpression(right)
	default:
		return newError("unknown operator: %s%s", operator, right.Type())
	}
}

func evalInfixExpression(
	operator string,
	left, right object.Object,
) object.Object {
	switch {
	case left.Type() == object.NUMBER_OBJ && right.Type() == object.NUMBER_OBJ:
		return evalNUMBERInfixExpression(operator, left, right)
	case left.Type() == object.RASTER_OBJ && right.Type() == object.NUMBER_OBJ:
		return evalRASTERNUMBERInfixExpression(operator, left, right)
	case left.Type() == object.NUMBER_OBJ && right.Type() == object.RASTER_OBJ:
		return evalNUMBERRASTERInfixExpression(operator, left, right)
	case left.Type() == object.RASTER_OBJ && right.Type() == object.RASTER_OBJ:
		return evalRASTERInfixExpression(operator, left, right)
	case operator == "==":
		return nativeBoolToBooleanObject(left == right)
	case operator == "!=":
		return nativeBoolToBooleanObject(left != right)
	case left.Type() != right.Type():
		return newError("type mismatch: %s %s %s",
			left.Type(), operator, right.Type())
	default:
		return newError("unknown operator: %s %s %s",
			left.Type(), operator, right.Type())
	}
}

func evalBangOperatorExpression(right object.Object) object.Object {
	switch right {
	case TRUE:
		return FALSE
	case FALSE:
		return TRUE
	case NULL:
		return TRUE
	default:
		return FALSE
	}
}

func evalMinusPrefixOperatorExpression(right object.Object) object.Object {
	if right.Type() != object.NUMBER_OBJ {
		return newError("unknown operator: -%s", right.Type())
	}

	value := right.(*object.Number).Value
	return &object.Number{Value: -value}
}

func evalNUMBERInfixExpression(operator string, left, right object.Object) object.Object {
	leftVal := left.(*object.Number).Value
	rightVal := right.(*object.Number).Value

	switch operator {
	case "+":
		return &object.Number{Value: leftVal + rightVal}
	case "-":
		return &object.Number{Value: leftVal - rightVal}
	case "*":
		return &object.Number{Value: leftVal * rightVal}
	case "/":
		return &object.Number{Value: leftVal / rightVal}
	case "<":
		return nativeBoolToBooleanObject(leftVal < rightVal)
	case ">":
		return nativeBoolToBooleanObject(leftVal > rightVal)
	case "==":
		return nativeBoolToBooleanObject(leftVal == rightVal)
	case "!=":
		return nativeBoolToBooleanObject(leftVal != rightVal)
	default:
		return newError("unknown operator: %s %s %s",
			left.Type(), operator, right.Type())
	}
}

func evalRASTERNUMBERInfixExpression(operator string, left, right object.Object) object.Object {
	rightVal := right.(*object.Number).Value
	leftVal := left.(*object.Raster).Value

	canvas := make([]float32, leftVal.Width*leftVal.Height)
	switch operator {
	case "+":
		for i, val := range leftVal.Data {
			canvas[i] = val + rightVal
		}
		return &object.Raster{Value: raster.FlexRaster{leftVal.RasterType, int(leftVal.Width), int(leftVal.Height), canvas, leftVal.NoData}}
	case "-":
		for i, val := range leftVal.Data {
			canvas[i] = val - rightVal
		}
		return &object.Raster{Value: raster.FlexRaster{leftVal.RasterType, int(leftVal.Width), int(leftVal.Height), canvas, leftVal.NoData}}
	case "*":
		for i, val := range leftVal.Data {
			canvas[i] = val * rightVal
		}
		return &object.Raster{Value: raster.FlexRaster{leftVal.RasterType, int(leftVal.Width), int(leftVal.Height), canvas, leftVal.NoData}}
	case "/":
		for i, val := range leftVal.Data {
			canvas[i] = val / rightVal
		}
		return &object.Raster{Value: raster.FlexRaster{leftVal.RasterType, int(leftVal.Width), int(leftVal.Height), canvas, leftVal.NoData}}
	case "==":
		for i, val := range leftVal.Data {
			if val == rightVal {
				canvas[i] = 1.0
			} else {
				canvas[i] = 0.0
			}
		}
		return &object.Raster{Value: raster.FlexRaster{raster.BOOL, int(leftVal.Width), int(leftVal.Height), canvas, leftVal.NoData}}
	default:
		return newError("unknown operator: %s %s %s",
			left.Type(), operator, right.Type())
	}
}

func evalNUMBERRASTERInfixExpression(operator string, left, right object.Object) object.Object {
	rightVal := right.(*object.Raster).Value
	leftVal := left.(*object.Number).Value

	canvas := make([]float32, rightVal.Width*rightVal.Height)
	switch operator {
	case "+":
		for i, val := range rightVal.Data {
			canvas[i] = val + leftVal
		}
		return &object.Raster{Value: raster.FlexRaster{rightVal.RasterType, int(rightVal.Width), int(rightVal.Height), canvas, rightVal.NoData}}
	case "-":
		for i, val := range rightVal.Data {
			canvas[i] = val - leftVal
		}
		return &object.Raster{Value: raster.FlexRaster{rightVal.RasterType, int(rightVal.Width), int(rightVal.Height), canvas, rightVal.NoData}}
	case "*":
		for i, val := range rightVal.Data {
			canvas[i] = val * leftVal
		}
		return &object.Raster{Value: raster.FlexRaster{rightVal.RasterType, int(rightVal.Width), int(rightVal.Height), canvas, rightVal.NoData}}
	case "/":
		for i, val := range rightVal.Data {
			canvas[i] = val / leftVal
		}
		return &object.Raster{Value: raster.FlexRaster{rightVal.RasterType, int(rightVal.Width), int(rightVal.Height), canvas, rightVal.NoData}}
	case "==":
		for i, val := range rightVal.Data {
			if val == leftVal {
				canvas[i] = 1.0
			} else {
				canvas[i] = 0.0
			}
		}
		return &object.Raster{Value: raster.FlexRaster{raster.BOOL, int(rightVal.Width), int(rightVal.Height), canvas, rightVal.NoData}}
	default:
		return newError("unknown operator: %s %s %s",
			left.Type(), operator, right.Type())
	}
}

func evalRASTERInfixExpression(operator string, left, right object.Object) object.Object {
	leftVal := left.(*object.Raster).Value
	rightVal := right.(*object.Raster).Value

	if leftVal.Width != rightVal.Width || leftVal.Height != rightVal.Height {
		return newError("non compatible rasters: Different width/height dimensions found.", leftVal.Width, leftVal.Height, rightVal.Width, rightVal.Height)
	}

	if len(leftVal.Data) != len(rightVal.Data) {
		return newError("non compatible rasters: Different data dimensions found: %d and %d", len(leftVal.Data), len(rightVal.Data))
	}

	canvas := make([]float32, leftVal.Width*leftVal.Height)

	switch operator {
	case "#":
		fmt.Println("AAAAAAAAAAAAAAA")
		if rightVal.RasterType != raster.BOOL {
			return newError("Raster on the right must be a Boolean raster type.")
		}
		for i, val := range rightVal.Data {
			if val == 1.0 {
				canvas[i] = leftVal.NoData
			} else {
				fmt.Println("AAAA", leftVal.Data[i])
				canvas[i] = leftVal.Data[i]
			}
		}
		return &object.Raster{Value: raster.FlexRaster{leftVal.RasterType, int(leftVal.Width), int(leftVal.Height), canvas, leftVal.NoData}}
	}

	if leftVal.RasterType != rightVal.RasterType {
		return newError("non compatible rasters: Different RasterType values found.")
	}

	if leftVal.NoData != rightVal.NoData {
		return newError("non compatible rasters: Different NoData values found.")
	}

	switch operator {
	case "+":
		for i, val := range leftVal.Data {
			canvas[i] = val + rightVal.Data[i]
		}
		return &object.Raster{Value: raster.FlexRaster{leftVal.RasterType, int(leftVal.Width), int(leftVal.Height), canvas, leftVal.NoData}}
	case "-":
		for i, val := range leftVal.Data {
			canvas[i] = val - rightVal.Data[i]
		}
		return &object.Raster{Value: raster.FlexRaster{leftVal.RasterType, int(leftVal.Width), int(leftVal.Height), canvas, leftVal.NoData}}
	case "*":
		for i, val := range leftVal.Data {
			canvas[i] = val * rightVal.Data[i]
		}
		return &object.Raster{Value: raster.FlexRaster{leftVal.RasterType, int(leftVal.Width), int(leftVal.Height), canvas, leftVal.NoData}}
	case "/":
		for i, val := range leftVal.Data {
			canvas[i] = val / rightVal.Data[i]
		}
		return &object.Raster{Value: raster.FlexRaster{leftVal.RasterType, int(leftVal.Width), int(leftVal.Height), canvas, leftVal.NoData}}
	}

	return newError("unknown operator: %s %s %s",
		left.Type(), operator, right.Type())
}

func evalIdentifier(node *ast.Identifier, env *object.Environment) object.Object {
	r, err := raster.GetRaster(node.Value)
	if err != nil {
		return newError("Raster reading operation failed")
	}
	return &object.Raster{Value: *r}
}

func isTruthy(obj object.Object) bool {
	switch obj {
	case NULL:
		return false
	case TRUE:
		return true
	case FALSE:
		return false
	default:
		return true
	}
}

func newError(format string, a ...interface{}) *object.Error {
	return &object.Error{Message: fmt.Sprintf(format, a...)}
}

func isError(obj object.Object) bool {
	if obj != nil {
		return obj.Type() == object.ERROR_OBJ
	}
	return false
}

func evalExpressions(exps []ast.Expression, env *object.Environment) []object.Object {
	var result []object.Object

	for _, e := range exps {
		evaluated := Eval(e, env)
		if isError(evaluated) {
			return []object.Object{evaluated}
		}
		result = append(result, evaluated)
	}

	return result
}
