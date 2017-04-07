package evaluator

import (
	"../ast"
	"../object"
	"../raster"
	"fmt"
	"image"
	"image/color"
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

	switch leftVal := left.(*object.Raster).Value.(type) {
	case *image.Gray:
		canvas := image.NewGray(leftVal.Bounds())
		b := leftVal.Bounds()
		switch operator {
		case "+":
			for y := b.Min.Y; y < b.Max.Y; y++ {
				for x := b.Min.X; x < b.Max.X; x++ {
					canvas.Set(x, y, color.Gray{Y: uint8(float64(leftVal.GrayAt(x, y).Y) + rightVal)})
				}
			}
			return &object.Raster{Value: canvas}
		case "-":
			for y := b.Min.Y; y < b.Max.Y; y++ {
				for x := b.Min.X; x < b.Max.X; x++ {
					canvas.Set(x, y, color.Gray{Y: uint8(float64(leftVal.GrayAt(x, y).Y) - rightVal)})
				}
			}
			return &object.Raster{Value: canvas}
		case "*":
			for y := b.Min.Y; y < b.Max.Y; y++ {
				for x := b.Min.X; x < b.Max.X; x++ {
					canvas.Set(x, y, color.Gray{Y: uint8(float64(leftVal.GrayAt(x, y).Y) * rightVal)})
				}
			}
			return &object.Raster{Value: canvas}
		case "/":
			for y := b.Min.Y; y < b.Max.Y; y++ {
				for x := b.Min.X; x < b.Max.X; x++ {
					canvas.Set(x, y, color.Gray{Y: uint8(float64(leftVal.GrayAt(x, y).Y) / rightVal)})
				}
			}
			return &object.Raster{Value: canvas}
		default:
			return newError("unknown operator: %s %s %s",
				left.Type(), operator, right.Type())
		}
	case *image.Gray16:
		canvas := image.NewGray16(leftVal.Bounds())
		b := leftVal.Bounds()
		switch operator {
		case "+":
			for y := b.Min.Y; y < b.Max.Y; y++ {
				for x := b.Min.X; x < b.Max.X; x++ {
					canvas.Set(x, y, color.Gray16{Y: uint16(float64(leftVal.Gray16At(x, y).Y) + rightVal)})
				}
			}
			return &object.Raster{Value: canvas}
		case "-":
			for y := b.Min.Y; y < b.Max.Y; y++ {
				for x := b.Min.X; x < b.Max.X; x++ {
					canvas.Set(x, y, color.Gray16{Y: uint16(float64(leftVal.Gray16At(x, y).Y) - rightVal)})
				}
			}
			return &object.Raster{Value: canvas}
		case "*":
			for y := b.Min.Y; y < b.Max.Y; y++ {
				for x := b.Min.X; x < b.Max.X; x++ {
					canvas.Set(x, y, color.Gray16{Y: uint16(float64(leftVal.Gray16At(x, y).Y) * rightVal)})
				}
			}
			return &object.Raster{Value: canvas}
		case "/":
			for y := b.Min.Y; y < b.Max.Y; y++ {
				for x := b.Min.X; x < b.Max.X; x++ {
					canvas.Set(x, y, color.Gray16{Y: uint16(float64(leftVal.Gray16At(x, y).Y) / rightVal)})
				}
			}
			return &object.Raster{Value: canvas}
		default:
			return newError("unknown operator: %s %s %s",
				left.Type(), operator, right.Type())
		}
		return &object.Raster{Value: canvas}
	default:
		return newError("unknown raster type: %s %s %s",
			left.Type(), operator, right.Type())
	}
}

func evalNUMBERRASTERInfixExpression(operator string, left, right object.Object) object.Object {
	leftVal := left.(*object.Number).Value

	switch rightVal := right.(*object.Raster).Value.(type) {
	case *image.Gray:
		canvas := image.NewGray(rightVal.Bounds())
		b := rightVal.Bounds()
		switch operator {
		case "+":
			for y := b.Min.Y; y < b.Max.Y; y++ {
				for x := b.Min.X; x < b.Max.X; x++ {
					canvas.Set(x, y, color.Gray{Y: uint8(float64(rightVal.GrayAt(x, y).Y) + leftVal)})
				}
			}
			return &object.Raster{Value: canvas}
		case "-":
			for y := b.Min.Y; y < b.Max.Y; y++ {
				for x := b.Min.X; x < b.Max.X; x++ {
					canvas.Set(x, y, color.Gray{Y: uint8(float64(rightVal.GrayAt(x, y).Y) - leftVal)})
				}
			}
			return &object.Raster{Value: canvas}
		case "*":
			for y := b.Min.Y; y < b.Max.Y; y++ {
				for x := b.Min.X; x < b.Max.X; x++ {
					canvas.Set(x, y, color.Gray{Y: uint8(float64(rightVal.GrayAt(x, y).Y) * leftVal)})
				}
			}
			return &object.Raster{Value: canvas}
		case "/":
			for y := b.Min.Y; y < b.Max.Y; y++ {
				for x := b.Min.X; x < b.Max.X; x++ {
					canvas.Set(x, y, color.Gray{Y: uint8(float64(rightVal.GrayAt(x, y).Y) / leftVal)})
				}
			}
			return &object.Raster{Value: canvas}
		default:
			return newError("unknown operator: %s %s %s",
				left.Type(), operator, right.Type())
		}
	case *image.Gray16:
		canvas := image.NewGray16(rightVal.Bounds())
		b := rightVal.Bounds()
		switch operator {
		case "+":
			for y := b.Min.Y; y < b.Max.Y; y++ {
				for x := b.Min.X; x < b.Max.X; x++ {
					canvas.Set(x, y, color.Gray16{Y: uint16(float64(rightVal.Gray16At(x, y).Y) + leftVal)})
				}
			}
			return &object.Raster{Value: canvas}
		case "-":
			for y := b.Min.Y; y < b.Max.Y; y++ {
				for x := b.Min.X; x < b.Max.X; x++ {
					canvas.Set(x, y, color.Gray16{Y: uint16(float64(rightVal.Gray16At(x, y).Y) - leftVal)})
				}
			}
			return &object.Raster{Value: canvas}
		case "*":
			for y := b.Min.Y; y < b.Max.Y; y++ {
				for x := b.Min.X; x < b.Max.X; x++ {
					canvas.Set(x, y, color.Gray16{Y: uint16(float64(rightVal.Gray16At(x, y).Y) * leftVal)})
				}
			}
			return &object.Raster{Value: canvas}
		case "/":
			for y := b.Min.Y; y < b.Max.Y; y++ {
				for x := b.Min.X; x < b.Max.X; x++ {
					canvas.Set(x, y, color.Gray16{Y: uint16(float64(rightVal.Gray16At(x, y).Y) / leftVal)})
				}
			}
			return &object.Raster{Value: canvas}
		default:
			return newError("unknown operator: %s %s %s",
				left.Type(), operator, right.Type())
		}
		return &object.Raster{Value: canvas}
	default:
		return newError("unknown raster type: %s %s %s",
			left.Type(), operator, right.Type())
	}
}

func evalRASTERInfixExpression(operator string, left, right object.Object) object.Object {

	leftVal, okLeft := left.(*object.Raster).Value.(*image.Gray)
	rightVal, okRight := right.(*object.Raster).Value.(*image.Gray)
	if okLeft && okRight {
		canvas := image.NewGray(rightVal.Bounds())
		b := rightVal.Bounds()
		switch operator {
		case "+":
			for y := b.Min.Y; y < b.Max.Y; y++ {
				for x := b.Min.X; x < b.Max.X; x++ {
					canvas.Set(x, y, color.Gray{Y: leftVal.GrayAt(x, y).Y + rightVal.GrayAt(x, y).Y})
				}
			}
			return &object.Raster{Value: canvas}
		case "-":
			for y := b.Min.Y; y < b.Max.Y; y++ {
				for x := b.Min.X; x < b.Max.X; x++ {
					canvas.Set(x, y, color.Gray{Y: leftVal.GrayAt(x, y).Y - rightVal.GrayAt(x, y).Y})
				}
			}
			return &object.Raster{Value: canvas}
		case "*":
			for y := b.Min.Y; y < b.Max.Y; y++ {
				for x := b.Min.X; x < b.Max.X; x++ {
					canvas.Set(x, y, color.Gray{Y: leftVal.GrayAt(x, y).Y * rightVal.GrayAt(x, y).Y})
				}
			}
			return &object.Raster{Value: canvas}
		case "/":
			for y := b.Min.Y; y < b.Max.Y; y++ {
				for x := b.Min.X; x < b.Max.X; x++ {
					if rightVal.GrayAt(x, y).Y != 0 {
						canvas.Set(x, y, color.Gray{Y: leftVal.GrayAt(x, y).Y / rightVal.GrayAt(x, y).Y})
					}
				}
			}
			return &object.Raster{Value: canvas}
		default:
			return newError("unknown operator: %s %s %s",
				left.Type(), operator, right.Type())
		}
	}

	left16Val, okLeft := left.(*object.Raster).Value.(*image.Gray16)
	right16Val, okRight := right.(*object.Raster).Value.(*image.Gray16)
	if okLeft && okRight {
		canvas := image.NewGray16(right16Val.Bounds())
		b := right16Val.Bounds()
		switch operator {
		case "+":
			for y := b.Min.Y; y < b.Max.Y; y++ {
				for x := b.Min.X; x < b.Max.X; x++ {
					canvas.Set(x, y, color.Gray16{Y: left16Val.Gray16At(x, y).Y + right16Val.Gray16At(x, y).Y})
				}
			}
			return &object.Raster{Value: canvas}
		case "-":
			for y := b.Min.Y; y < b.Max.Y; y++ {
				for x := b.Min.X; x < b.Max.X; x++ {
					canvas.Set(x, y, color.Gray16{Y: left16Val.Gray16At(x, y).Y - right16Val.Gray16At(x, y).Y})
				}
			}
			return &object.Raster{Value: canvas}
		case "*":
			for y := b.Min.Y; y < b.Max.Y; y++ {
				for x := b.Min.X; x < b.Max.X; x++ {
					canvas.Set(x, y, color.Gray16{Y: left16Val.Gray16At(x, y).Y * right16Val.Gray16At(x, y).Y})
				}
			}
			return &object.Raster{Value: canvas}
		case "/":
			for y := b.Min.Y; y < b.Max.Y; y++ {
				for x := b.Min.X; x < b.Max.X; x++ {
					if right16Val.Gray16At(x, y).Y != 0 {
						canvas.Set(x, y, color.Gray16{Y: left16Val.Gray16At(x, y).Y / right16Val.Gray16At(x, y).Y})
					}
				}
			}
			return &object.Raster{Value: canvas}
		default:
			return newError("unknown operator: %s %s %s",
				left.Type(), operator, right.Type())
		}
	}
	return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
}

func evalIdentifier(node *ast.Identifier, env *object.Environment) object.Object {
	im, err := raster.GetRaster(node.Value)
	if err != nil {
		return newError("Raster reading operation failed")
	}
	return &object.Raster{Value: im}
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
