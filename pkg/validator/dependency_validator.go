package validator

import "PROJECT_NAME/pkg/errs"

// DependencyValidator validates if dependencies are been correctly provided
type DependencyValidator struct {
	context string
	deps    map[string]any
}

// NewDependencyValidator creates a new validator instance
func NewDependencyValidator(context string) *DependencyValidator {
	return &DependencyValidator{
		context: context,
		deps:    make(map[string]any),
	}
}

// Check adds a new dependency for validation and returns the validator instance
func (dv *DependencyValidator) Check(depDesc string, dep any) *DependencyValidator {
	dv.deps[depDesc] = dep
	return dv
}

func (dv *DependencyValidator) UseCaseCheck(ucDesc string, dep any) *DependencyValidator {
	desc := "Caso de uso: " + ucDesc
	return dv.Check(desc, dep)
}

func (dv *DependencyValidator) RepositoryCheck(repoDesc string, dep any) *DependencyValidator {
	desc := "Repositório: " + repoDesc
	return dv.Check(desc, dep)
}

func (dv *DependencyValidator) QueryPortCheck(repoDesc string, dep any) *DependencyValidator {
	desc := "Porta de consulta: " + repoDesc
	return dv.Check(desc, dep)
}

func (dv *DependencyValidator) ServiceCheck(svcDesc string, dep any) *DependencyValidator {
	desc := "Serviço: " + svcDesc
	return dv.Check(desc, dep)
}

func (dv *DependencyValidator) ApplicationCheck(appDesc string, dep any) *DependencyValidator {
	desc := "Aplicação: " + appDesc
	return dv.Check(desc, dep)
}

func (dv *DependencyValidator) DatabaseCheck(dbName string, dep any) *DependencyValidator {
	desc := "Banco de Dados: " + dbName
	return dv.Check(desc, dep)
}

// MustValidate run's the validation and fires a panic if one or more dependencies is nil
func (dv *DependencyValidator) MustValidate() {
	for desc, dep := range dv.deps {
		if dep == nil {
			panic(errs.ErrMissingRequiredDependency(desc, dv.context).Error())
		}
	}
}
